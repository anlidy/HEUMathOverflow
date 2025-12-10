package response

// ES搜索响应
type PostMeta struct {
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}
type SearchResponse struct {
	Total     int
	PostMetas []PostMeta
}
