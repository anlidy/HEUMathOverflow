package request

type PostCreate struct {
	ClientToken string   `json:"client_token" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	ImageURLs   []string `json:"image_urls"`
	Tags        []string `json:"tags"`
}

type PostUpdate struct {
	Title           string   `json:"title" binding:"required"`
	Content         string   `json:"content" binding:"required"`
	AddImageURLs    []string `json:"add_image_urls"`
	DeleteImageURLs []string `json:"delete_image_urls"`
	Tags            []string `json:"tags"`
}

type PostCertifiedUpdate struct {
	IsCertified bool `json:"is_certified"`
}
