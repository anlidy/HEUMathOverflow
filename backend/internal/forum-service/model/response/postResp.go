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
	Liked   bool `json:"liked"`
	Starred bool `json:"starred"`
}

type MultiPostData struct {
	UserInfo UserInfo `json:"user_info"`
	PostData PostData `json:"post_data"`
}
