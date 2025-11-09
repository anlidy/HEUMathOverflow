package response

import (
	"MathOverflow/internal/forum-service/model"
	"time"
)

type UserInfo struct {
	ID        int64  `json:"user_id,string"`
	Username  string `json:"username"`
	Role      int    `json:"role,string"`
	AvatarUrl string `json:"avatar_url"`
}

type PostData struct {
	ID          int64            `json:"post_id,string"`
	Title       string           `json:"title"`
	Tags        []string         `json:"tags"`
	Status      model.PostStatus `json:"status,string"`
	Content     string           `json:"content"`
	Images      []string         `json:"images"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	LastReplyAt *time.Time       `json:"last_reply_at"`
}
