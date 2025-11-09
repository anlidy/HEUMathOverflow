package response

import (
	"MathOverflow/internal/forum-service/model"
	"time"
)

type ReplyData struct {
	ID            int64             `json:"reply_id,string"`
	PostID        int64             `json:"post_id,string"`
	ParentReplyID *int64            `json:"parent_reply_id,string"`
	Status        model.ReplyStatus `json:"status,string"`
	CertifiedBy   *int64            `json:"certified_by,string"`
	Content       string            `json:"content"`
	Voice         string            `json:"voice"` //存储语音url
	VoiceText     string            `json:"voice_text"`
	Images        []string          `json:"images"`
	AIAnswered    bool              `json:"ai_answered"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
