package response

import (
	"MathOverflow/services/forum/internal/model"
	"encoding/json"
	"time"
)

type ChatSessionItem struct {
	SessionID            string    `json:"session_id"`
	Title                string    `json:"title"`
	Status               string    `json:"status"`
	Summary              string    `json:"summary"`
	SummaryUpToMessageID *int64    `json:"summary_up_to_message_id,string,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ChatMessageItem struct {
	MessageID      int64           `json:"message_id,string"`
	SessionID      string          `json:"session_id"`
	TurnID         string          `json:"turn_id"`
	Role           string          `json:"role"`
	Type           string          `json:"type"`
	Content        string          `json:"content"`
	ToolName       string          `json:"tool_name,omitempty"`
	ToolCallID     string          `json:"tool_call_id,omitempty"`
	ToolArgsJSON   string          `json:"tool_args_json,omitempty"`
	ToolResultJSON string          `json:"tool_result_json,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type ChatToolTrace struct {
	ToolName    string          `json:"tool_name"`
	ToolCallID  string          `json:"tool_call_id"`
	ToolArgs    json.RawMessage `json:"tool_args,omitempty"`
	ToolResult  json.RawMessage `json:"tool_result,omitempty"`
	Content     string          `json:"content"`
	Status      string          `json:"status"`
}

type ChatTurnData struct {
	Session         ChatSessionItem   `json:"session"`
	UserMessage     ChatMessageItem   `json:"user_message"`
	AssistantMessage ChatMessageItem  `json:"assistant_message"`
	ToolTraces      []ChatToolTrace   `json:"tool_traces"`
	UsedTools       []string          `json:"used_tools"`
	RetrievedPosts  json.RawMessage   `json:"retrieved_posts,omitempty"`
	SummaryUpdated  bool              `json:"summary_updated"`
}

func NewChatSessionItem(session model.ChatSession) ChatSessionItem {
	return ChatSessionItem{
		SessionID:            session.ID,
		Title:                session.Title,
		Status:               string(session.Status),
		Summary:              session.Summary,
		SummaryUpToMessageID: session.SummaryUpToMessageID,
		CreatedAt:            session.CreatedAt,
		UpdatedAt:            session.UpdatedAt,
	}
}

func NewChatMessageItem(message model.ChatMessage) ChatMessageItem {
	return ChatMessageItem{
		MessageID:      message.ID,
		SessionID:      message.SessionID,
		TurnID:         message.TurnID,
		Role:           string(message.Role),
		Type:           string(message.MessageType),
		Content:        message.Content,
		ToolName:       message.ToolName,
		ToolCallID:     message.ToolCallID,
		ToolArgsJSON:   message.ToolArgsJSON,
		ToolResultJSON: message.ToolResultJSON,
		CreatedAt:      message.CreatedAt,
	}
}
