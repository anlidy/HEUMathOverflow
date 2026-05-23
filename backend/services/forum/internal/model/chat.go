package model

import "time"

type ChatSessionStatus string

const (
	ChatSessionActive ChatSessionStatus = "active"
)

type ChatMessageRole string

const (
	ChatRoleSystem    ChatMessageRole = "system"
	ChatRoleUser      ChatMessageRole = "user"
	ChatRoleAssistant ChatMessageRole = "assistant"
	ChatRoleTool      ChatMessageRole = "tool"
)

type ChatMessageType string

const (
	ChatMessageText       ChatMessageType = "text"
	ChatMessageToolUse    ChatMessageType = "tool_use"
	ChatMessageToolResult ChatMessageType = "tool_result"
	ChatMessageEvent      ChatMessageType = "event"
)

type ChatSession struct {
	ID                   string            `gorm:"primaryKey;size:64" json:"session_id"`
	UserID               int64             `gorm:"not null;index" json:"user_id,string"`
	Title                string            `gorm:"size:255" json:"title"`
	Status               ChatSessionStatus `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
	Summary              string            `gorm:"type:text" json:"summary"`
	SummaryUpToMessageID *int64            `json:"summary_up_to_message_id,string"`
	CreatedAt            time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

type ChatMessage struct {
	ID             int64           `gorm:"primaryKey;autoIncrement:false" json:"message_id,string"`
	SessionID      string          `gorm:"size:64;not null;index" json:"session_id"`
	TurnID         string          `gorm:"size:64;not null;index" json:"turn_id"`
	Role           ChatMessageRole `gorm:"type:varchar(16);not null" json:"role"`
	MessageType    ChatMessageType `gorm:"column:message_type;type:varchar(16);not null" json:"type"`
	Content        string          `gorm:"type:text" json:"content"`
	ToolName       string          `gorm:"size:128" json:"tool_name,omitempty"`
	ToolCallID     string          `gorm:"size:128;index" json:"tool_call_id,omitempty"`
	ToolArgsJSON   string          `gorm:"type:text" json:"tool_args_json,omitempty"`
	ToolResultJSON string          `gorm:"type:text" json:"tool_result_json,omitempty"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
}
