package request

type ReplyCreate struct {
	PostID        int64  `json:"post_id,string"`
	ParentReplyID *int64 `json:"parent_reply_id,string"`
	/// mongo
	Content   string   `json:"content"`
	VoiceURL  string   `json:"voice_url"`
	ImageURLs []string `bson:"image_urls"`
}
