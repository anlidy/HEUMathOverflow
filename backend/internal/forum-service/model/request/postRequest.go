package request

type PostCreate struct {
	Title    string   `json:"title"`
	Content  string   `json:"content" binding:"required"`
	ImageURL []string `json:"image_url"`
	Tags     []string `json:"tags"`
}
