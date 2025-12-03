package request

type ReplyCreate struct {
	PostID        int64    `json:"post_id,string" binding:"required"`
	ParentReplyID *int64   `json:"parent_reply_id,string"`
	Content       string   `json:"content" binding:"required"`
	VoiceURL      string   `json:"voice_url"`
	ImageURLs     []string `bson:"image_urls"`
}

type ReplyUpdate struct {
	Content         string   `json:"content" binding:"required"`
	VoiceURL        string   `json:"voice_url"` // 不一定有,不设置binding
	AddImageURLs    []string `json:"add_image_urls"`
	DeleteImageURLs []string `json:"delete_image_urls"`
}
