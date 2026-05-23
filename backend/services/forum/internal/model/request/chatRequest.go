package request

type ChatSessionCreate struct {
	Title string `json:"title"`
}

type ChatMessageCreate struct {
	Content string `json:"content" binding:"required"`
}
