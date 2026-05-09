package event

import (
	"MathOverflow/common/client"
	"MathOverflow/common/utils"
	"fmt"
	"time"
)

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

func PublishRagEvent(mq *client.RabbitMQClient, t RagEventType, payload RagDataPayload) {
	if mq == nil {
		utils.Logger().Warn("mq is nil, cannot publish post event")
		return
	}

	evt := RagDataEvent{
		Type:      t,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	// 计算分片
	if mq.RagWorkerCount <= 0 {
		utils.Logger().Warn("rag_worker_count must >= 0")
	}
	shardID := payload.PostID % mq.RagWorkerCount
	routeKey := fmt.Sprintf("%s.%d", t, shardID) // rag.data.created.1

	if err := mq.PublishEvent(routeKey, evt); err != nil {
		utils.Logger().WithError(err).Error("publish rag event failed")
	}
}
