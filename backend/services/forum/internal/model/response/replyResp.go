package response

import (
	"MathOverflow/services/forum/internal/model"
)

type ReplyData struct {
	model.Reply
	Liked bool `json:"liked"`
}

type MultiReplyData struct {
	UserInfo  UserInfo  `json:"user_info"`
	ReplyData ReplyData `json:"reply_data"`
}
