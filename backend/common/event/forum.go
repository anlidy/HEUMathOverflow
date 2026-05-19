package event

import "time"

// ForumEventType defines post domain event types.
type ForumEventType string

const (
	ForumPostCreated     ForumEventType = "forum.post.created"
	ForumPostUpdated     ForumEventType = "forum.post.updated"
	ForumPostDeleted     ForumEventType = "forum.post.deleted"
	ForumPostStatUpdated ForumEventType = "forum.post.statUpdated"
)

// ForumPostPayload is the transport-friendly shape for posts.
type ForumPostPayload struct {
	PostID     int64     `json:"post_id,omitempty"` // post_id
	AuthorID   int64     `json:"author_id,omitempty"`
	AuthorName string    `json:"author_name,omitempty"`
	Title      string    `json:"title,omitempty"`
	Content    string    `json:"content,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	Status     int       `json:"status,omitempty"`
	Views      int64     `json:"views"`   // 浏览量
	Likes      int64     `json:"likes"`   // 赞同数
	Stars      int64     `json:"stars"`   // 收藏数
	Replies    int64     `json:"replies"` // 评论数
	CreatedAt  time.Time `json:"created_at,omitempty"`
	Favors     int64     `json:"favors"`
	TotalScore float64   `json:"total_score"`
}

// ForumPostEvent is the envelope delivered over MQ.
type ForumPostEvent struct {
	Type      ForumEventType   `json:"type"`
	Payload   ForumPostPayload `json:"payload"`
	CreatedAt time.Time        `json:"created_at"`
}
