package service

import (
	"MathOverflow/common/config"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/model"
	requestmodel "MathOverflow/services/forum/internal/model/request"
	responsemodel "MathOverflow/services/forum/internal/model/response"
	"MathOverflow/services/forum/internal/repository"
	bytespkg "bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	syserror "MathOverflow/services/forum/internal/model/error"

	"gorm.io/gorm"
)

type ChatService interface {
	CreateSession(ctx context.Context, userID int64, req requestmodel.ChatSessionCreate) (model.ChatSession, syserror.Error)
	ListSessions(ctx context.Context, userID int64, page, pageSize int) ([]model.ChatSession, int64, syserror.Error)
	ListMessages(ctx context.Context, userID int64, sessionID string, page, pageSize int) (model.ChatSession, []model.ChatMessage, int64, syserror.Error)
	SendMessage(ctx context.Context, userID int64, sessionID string, req requestmodel.ChatMessageCreate) (*responsemodel.ChatTurnData, syserror.Error)
}

type chatService struct {
	cfg        config.Config
	chatRepo   repository.ChatRepo
	httpClient *http.Client
	servName   string
}

type ragChatMessage struct {
	MessageID  *int64         `json:"message_id,omitempty"`
	Role       string         `json:"role"`
	Type       string         `json:"type"`
	Content    string         `json:"content"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolArgs   map[string]any `json:"tool_args,omitempty"`
	ToolResult map[string]any `json:"tool_result,omitempty"`
}

type ragChatRequest struct {
	SessionID            string           `json:"session_id"`
	Summary              string           `json:"summary,omitempty"`
	SummaryUpToMessageID *int64           `json:"summary_up_to_message_id,omitempty"`
	Messages             []ragChatMessage `json:"messages"`
}

type ragToolTrace struct {
	ToolName   string         `json:"tool_name"`
	ToolCallID string         `json:"tool_call_id"`
	ToolArgs   map[string]any `json:"tool_args"`
	ToolResult map[string]any `json:"tool_result"`
	Content    string         `json:"content"`
	Status     string         `json:"status"`
}

type ragSummaryUpdate struct {
	Summary     string `json:"summary"`
	UpToMessageID int64 `json:"up_to_message_id"`
}

type ragChatResponse struct {
	Answer        string            `json:"answer"`
	UsedTools     []string          `json:"used_tools"`
	RetrievedPosts json.RawMessage  `json:"retrieved_posts"`
	ToolTraces    []ragToolTrace    `json:"tool_traces"`
	SummaryUpdate *ragSummaryUpdate `json:"summary_update"`
}

func NewChatService(cfg config.Config, chatRepo repository.ChatRepo) ChatService {
	timeout := cfg.Rag.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	}
	return &chatService{
		cfg:      cfg,
		chatRepo: chatRepo,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
		servName: "Chat-Service",
	}
}

func (s *chatService) CreateSession(ctx context.Context, userID int64, req requestmodel.ChatSessionCreate) (model.ChatSession, syserror.Error) {
	session := model.ChatSession{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		Title:     strings.TrimSpace(req.Title),
		Status:    model.ChatSessionActive,
		Summary:   "",
	}
	if err := s.chatRepo.CreateSession(&session); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("create chat session failed")
		return model.ChatSession{}, syserror.InternalError
	}
	return session, syserror.NoError
}

func (s *chatService) ListSessions(ctx context.Context, userID int64, page, pageSize int) ([]model.ChatSession, int64, syserror.Error) {
	offset := max(page-1, 0) * pageSize
	sessions, total, err := s.chatRepo.ListSessionsByUser(userID, offset, pageSize)
	if err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("list chat sessions failed")
		return nil, 0, syserror.InternalError
	}
	return sessions, total, syserror.NoError
}

func (s *chatService) ListMessages(ctx context.Context, userID int64, sessionID string, page, pageSize int) (model.ChatSession, []model.ChatMessage, int64, syserror.Error) {
	session, err := s.chatRepo.GetSessionByIDAndUser(sessionID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.ChatSession{}, nil, 0, syserror.NotFoundError
		}
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("load chat session failed")
		return model.ChatSession{}, nil, 0, syserror.InternalError
	}
	offset := max(page-1, 0) * pageSize
	messages, total, listErr := s.chatRepo.ListMessages(sessionID, offset, pageSize)
	if listErr != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(listErr).Error("list chat messages failed")
		return model.ChatSession{}, nil, 0, syserror.InternalError
	}
	return session, messages, total, syserror.NoError
}

func (s *chatService) SendMessage(ctx context.Context, userID int64, sessionID string, req requestmodel.ChatMessageCreate) (*responsemodel.ChatTurnData, syserror.Error) {
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, syserror.ConflictError
	}
	session, err := s.chatRepo.GetSessionByIDAndUser(sessionID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, syserror.NotFoundError
		}
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("load chat session failed before send")
		return nil, syserror.InternalError
	}

	userMessageID := utils.GenerateSnowflakeID()
	turnID := utils.GenerateUUID()
	rawMessages, listErr := s.chatRepo.ListMessagesAfter(sessionID, session.SummaryUpToMessageID)
	if listErr != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(listErr).Error("load chat messages after summary failed")
		return nil, syserror.InternalError
	}
	rawMessages = append(rawMessages, model.ChatMessage{
		ID:          userMessageID,
		SessionID:   sessionID,
		TurnID:      turnID,
		Role:        model.ChatRoleUser,
		MessageType: model.ChatMessageText,
		Content:     content,
	})
		ragResp, ragErr := s.callRagChat(ctx, session, rawMessages)
		if ragErr != syserror.NoError {
			return nil, ragErr
		}

	userMessage := model.ChatMessage{
		ID:          userMessageID,
		SessionID:   sessionID,
		TurnID:      turnID,
		Role:        model.ChatRoleUser,
		MessageType: model.ChatMessageText,
		Content:     content,
	}
	assistantMessage := model.ChatMessage{
		ID:          utils.GenerateSnowflakeID(),
		SessionID:   sessionID,
		TurnID:      turnID,
		Role:        model.ChatRoleAssistant,
		MessageType: model.ChatMessageText,
		Content:     ragResp.Answer,
	}
	messagesToPersist := []model.ChatMessage{userMessage}
	toolTraces := make([]responsemodel.ChatToolTrace, 0, len(ragResp.ToolTraces))
	for _, trace := range ragResp.ToolTraces {
		argsJSON, _ := json.Marshal(trace.ToolArgs)
		resultJSON, _ := json.Marshal(trace.ToolResult)
		toolUseContent := fmt.Sprintf("[tool_use] %s", trace.ToolName)
		if len(argsJSON) > 0 {
			toolUseContent = fmt.Sprintf("[tool_use] %s args:%s", trace.ToolName, string(argsJSON))
		}
		messagesToPersist = append(messagesToPersist,
			model.ChatMessage{
				ID:           utils.GenerateSnowflakeID(),
				SessionID:    sessionID,
				TurnID:       turnID,
				Role:         model.ChatRoleAssistant,
				MessageType:  model.ChatMessageToolUse,
				Content:      toolUseContent,
				ToolName:     trace.ToolName,
				ToolCallID:   trace.ToolCallID,
				ToolArgsJSON: string(argsJSON),
			},
			model.ChatMessage{
				ID:             utils.GenerateSnowflakeID(),
				SessionID:      sessionID,
				TurnID:         turnID,
				Role:           model.ChatRoleTool,
				MessageType:    model.ChatMessageToolResult,
				Content:        trace.Content,
				ToolName:       trace.ToolName,
				ToolCallID:     trace.ToolCallID,
				ToolArgsJSON:   string(argsJSON),
				ToolResultJSON: string(resultJSON),
			},
		)
		toolTraces = append(toolTraces, responsemodel.ChatToolTrace{
			ToolName:   trace.ToolName,
			ToolCallID: trace.ToolCallID,
			ToolArgs:   mustRawJSON(argsJSON),
			ToolResult: mustRawJSON(resultJSON),
			Content:    trace.Content,
			Status:     trace.Status,
		})
	}
	messagesToPersist = append(messagesToPersist, assistantMessage)
	summaryUpdated := ragResp.SummaryUpdate != nil && strings.TrimSpace(ragResp.SummaryUpdate.Summary) != ""
	if err := s.chatRepo.RunInTx(ctx, func(tx *gorm.DB) error {
		if err := s.chatRepo.CreateMessagesTx(tx, messagesToPersist); err != nil {
			return err
		}
		if session.Title == "" {
			session.Title = s.defaultSessionTitle(content)
		}
		if summaryUpdated {
			session.Summary = ragResp.SummaryUpdate.Summary
			session.SummaryUpToMessageID = &ragResp.SummaryUpdate.UpToMessageID
		}
		return s.chatRepo.UpdateSessionTx(tx, &session)
	}); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("persist chat turn failed")
		return nil, syserror.InternalError
	}
	data := &responsemodel.ChatTurnData{
		Session:          responsemodel.NewChatSessionItem(session),
		UserMessage:      responsemodel.NewChatMessageItem(userMessage),
		AssistantMessage: responsemodel.NewChatMessageItem(assistantMessage),
		ToolTraces:       toolTraces,
		UsedTools:        ragResp.UsedTools,
		RetrievedPosts:   ragResp.RetrievedPosts,
		SummaryUpdated:   summaryUpdated,
	}
	return data, syserror.NoError
}

func (s *chatService) callRagChat(ctx context.Context, session model.ChatSession, messages []model.ChatMessage) (*ragChatResponse, syserror.Error) {
	payload := ragChatRequest{
		SessionID:            session.ID,
		Summary:              session.Summary,
		SummaryUpToMessageID: session.SummaryUpToMessageID,
		Messages:             make([]ragChatMessage, 0, len(messages)),
	}
	for _, message := range messages {
		payload.Messages = append(payload.Messages, ragChatMessage{
			MessageID:  &message.ID,
			Role:       string(message.Role),
			Type:       string(message.MessageType),
			Content:    message.Content,
			ToolName:   message.ToolName,
			ToolCallID: message.ToolCallID,
			ToolArgs:   parseJSONMap(message.ToolArgsJSON),
			ToolResult: parseJSONMap(message.ToolResultJSON),
		})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("marshal rag chat request failed")
		return nil, syserror.InternalError
	}
	baseURL := strings.TrimRight(s.cfg.Rag.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://rag-service:9091"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat", bytespkg.NewReader(body))
	if err != nil {
		return nil, syserror.InternalError
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("call rag chat failed")
		return nil, syserror.NetworkError
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, syserror.InternalError
	}
	if resp.StatusCode >= 400 {
		utils.WithContext(ctx).WithField("service", s.servName).WithField("status_code", resp.StatusCode).Error("rag chat returned failure")
		return nil, syserror.InternalError
	}
	var ragResp ragChatResponse
	if err := json.Unmarshal(respBody, &ragResp); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decode rag chat response failed")
		return nil, syserror.InternalError
	}
	return &ragResp, syserror.NoError
}

func (s *chatService) defaultSessionTitle(content string) string {
	content = strings.TrimSpace(content)
	if len(content) <= 60 {
		return content
	}
	return content[:60]
}

func parseJSONMap(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil
	}
	return value
}

func mustRawJSON(raw []byte) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	return json.RawMessage(raw)
}
