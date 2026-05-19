package outbox

import (
	"MathOverflow/common/client"
	common "MathOverflow/common/model"
	"MathOverflow/common/utils"
	"context"
	"time"

	"gorm.io/gorm"
)

type RouteKeyFunc func(message common.OutboxMessage, mq *client.RabbitMQClient) string

type MessageWorker struct {
	mq           *client.RabbitMQClient
	store        Store
	target       string
	name         string
	routeKeyFunc RouteKeyFunc
}

func NewMessageWorker(db *gorm.DB, mq *client.RabbitMQClient, target, name string, routeKeyFunc RouteKeyFunc) MessageWorker {
	return MessageWorker{
		mq:           mq,
		store:        NewStore(db),
		target:       target,
		name:         name,
		routeKeyFunc: routeKeyFunc,
	}
}

func (w *MessageWorker) Run(ctx context.Context, interval time.Duration, batchSize int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	utils.Logger().WithField("worker", w.name).WithField("interval", interval.String()).Info("outbox message worker started")
	for {
		select {
		case <-ticker.C:
			w.sendPendingMessages(ctx, batchSize)
		case <-ctx.Done():
			utils.Logger().WithField("worker", w.name).Info("outbox message worker stopped")
			return
		}
	}
}

func (w *MessageWorker) sendPendingMessages(ctx context.Context, batchSize int) {
	messages, err := w.store.TakePendingMessages(ctx, w.target, batchSize)
	if err != nil {
		utils.Logger().WithField("worker", w.name).WithError(err).Error("load pending outbox messages failed")
		return
	}
	for _, message := range messages {
		if err := w.sendOneMessage(message); err != nil {
			retryAfter := time.Now().Add(time.Duration(min(message.RetryCount+1, 10)) * time.Minute)
			_ = w.store.MarkMessageRetry(ctx, message.ID, err.Error(), retryAfter)
			continue
		}
		_ = w.store.MarkMessageSent(ctx, message.ID)
	}
}

func (w *MessageWorker) sendOneMessage(message common.OutboxMessage) error {
	if w.mq == nil {
		return nil
	}
	routeKey := message.Topic
	if w.routeKeyFunc != nil {
		routeKey = w.routeKeyFunc(message, w.mq)
	}
	return w.mq.PublishEvent(routeKey, message.Payload)
}
