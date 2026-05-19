package outbox

import (
	common "MathOverflow/common/model"
	"MathOverflow/common/utils"
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Store interface {
	AddMessageInTx(ctx context.Context, tx *gorm.DB, resourceType string, resourceID int64, topic, target string, payload any) error
	TakePendingMessages(ctx context.Context, target string, limit int) ([]common.OutboxMessage, error)
	MarkMessageSent(ctx context.Context, id int64) error
	MarkMessageRetry(ctx context.Context, id int64, lastError string, retryAt time.Time) error
}

type GormStore struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return &GormStore{db: db}
}

func (s *GormStore) AddMessageInTx(ctx context.Context, tx *gorm.DB, resourceType string, resourceID int64, topic, target string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := common.OutboxMessage{
		ID:           utils.GenerateSnowflakeID(),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Topic:        topic,
		Target:       target,
		Payload:      body,
		Status:       common.OutboxMessagePending,
		RetryAt:      time.Now(),
	}
	return tx.WithContext(ctx).Create(&msg).Error
}

func (s *GormStore) TakePendingMessages(ctx context.Context, target string, limit int) ([]common.OutboxMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	var messages []common.OutboxMessage
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			SELECT *
			FROM outbox_messages
			WHERE target = ?
			  AND status = ?
			  AND retry_at <= NOW()
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT ?
		`, target, common.OutboxMessagePending, limit).Scan(&messages).Error; err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}
		ids := make([]int64, 0, len(messages))
		for _, msg := range messages {
			ids = append(ids, msg.ID)
		}
		return tx.Model(&common.OutboxMessage{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"retry_count": gorm.Expr("retry_count + 1"),
				"retry_at":    time.Now().Add(30 * time.Second),
			}).Error
	})
	return messages, err
}

func (s *GormStore) MarkMessageSent(ctx context.Context, id int64) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&common.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":  common.OutboxMessageSent,
			"sent_at": &now,
		}).Error
}

func (s *GormStore) MarkMessageRetry(ctx context.Context, id int64, lastError string, retryAt time.Time) error {
	return s.db.WithContext(ctx).Model(&common.OutboxMessage{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_error": lastError,
			"retry_at":   retryAt,
		}).Error
}
