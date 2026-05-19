package common

import "time"

type OutboxMessageStatus string

const (
	OutboxMessagePending OutboxMessageStatus = "pending"
	OutboxMessageSent    OutboxMessageStatus = "sent"
)

type OutboxMessage struct {
	ID           int64               `gorm:"primaryKey;autoIncrement:false"`
	ResourceType string              `gorm:"size:64;not null;index:idx_outbox_pending,priority:1"`
	ResourceID   int64               `gorm:"not null"`
	Topic        string              `gorm:"size:128;not null"`
	Target       string              `gorm:"size:64;not null;index:idx_outbox_pending,priority:2"`
	Payload      []byte              `gorm:"type:jsonb;not null"`
	Status       OutboxMessageStatus `gorm:"size:32;not null;index:idx_outbox_pending,priority:3"`
	RetryCount   int                 `gorm:"not null;default:0"`
	LastError    string              `gorm:"type:text"`
	RetryAt      time.Time           `gorm:"not null;index:idx_outbox_pending,priority:4"`
	SentAt       *time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
