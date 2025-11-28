package response

import (
	"MathOverflow/internal/forum-service/model"
)

type UserInfo struct {
	ID        int64  `json:"user_id,string"`
	Username  string `json:"username"`
	Role      int    `json:"role"`
	AvatarUrl string `json:"avatar_url"`
}

type PostData struct {
	model.Post
	model.PostContent
}
