package event

import "time"

type RagEventType string

const (
	RagDataCreated = "rag.data.created"
	RagDataUpdated = "rag.data.updated"
	RagDataDeleted = "rag.data.deleted"
)

type PostAnswer struct {
	ReplyID   int64  `json:"reply_id"`
	ReplierID int64  `gorm:"not null" json:"-"`
	Content   string `gorm:"type:text" json:"content"`
}

type RagDataPayload struct {
	PostID   int64        `json:"post_id"` // post_id
	AuthorID int64        `json:"author_id"`
	Title    string       `json:"title"`
	Content  string       `json:"content"`
	Tags     []string     `json:"tags"`
	Answers  []PostAnswer `json:"answers"`
}

type RagDataEvent struct {
	Type      RagEventType   `json:"type"`
	Payload   RagDataPayload `json:"payload"`
	CreatedAt time.Time      `json:"created_at"`
}
