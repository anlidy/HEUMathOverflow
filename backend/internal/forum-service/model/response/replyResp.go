package response

import (
	"MathOverflow/internal/forum-service/model"
)

type ReplyData struct {
	model.Reply
}

type MultiReplyData struct {
	UserInfo  UserInfo  `json:"user_info"`
	ReplyData ReplyData `json:"reply_data"`
}
