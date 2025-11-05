package request

type PostCreate struct {
	Title   string   `json:"title"`
	Content string   `json:"content" binding:"required"`
	Images  []string `json:"images"`
	Tags    []string `json:"tags"`
}
