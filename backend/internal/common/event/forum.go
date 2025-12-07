package event

import "time"

const (
	// forum post domain events
	ForumPostSearch  = "forum.post.search"
	ForumPostCreated = "forum.post.created"
	ForumPostUpdated = "forum.post.updated"
	ForumPostDeleted = "forum.post.deleted"
)

// ForumEventType defines post domain event types.
type ForumEventType string

const (
	PostCreated ForumEventType = ForumPostCreated
	PostUpdated                = ForumPostUpdated
	PostDeleted                = ForumPostDeleted
)

// ForumPostPayload is the transport-friendly shape for posts.
type ForumPostPayload struct {
	ID        int64     `json:"id,string"`
	AuthorID  int64     `json:"author_id,string"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ImageURLs []string  `json:"image_urls"`
	Tags      []string  `json:"tags"`
	Status    int       `json:"status"`
	Views     uint      `json:"views"`
	Likes     uint      `json:"likes"`
	Stars     uint      `json:"stars"`
	Replies   uint      `json:"replies"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ForumPostEvent is the envelope delivered over MQ.
type ForumPostEvent struct {
	Type      ForumEventType   `json:"type"`
	Payload   ForumPostPayload `json:"payload"`
	CreatedAt time.Time        `json:"created_at"`
}
