package consumer

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/event"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/model"
	"MathOverflow/internal/forum-service/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartUserConsumer starts a background consumer for user events and keeps running
// until ctx is canceled.
func StartUserConsumer(ctx context.Context, mq *client.RabbitMQClient, userRepo repository.UserSnapshotRepo) error {
	go startUserWorker(ctx, mq, userRepo)
	utils.Logger().WithField("service", "forum-service").Info("[UserConsumer] worker started")
	<-ctx.Done()
	return nil
}

func handleUserEvent(userRepo repository.UserSnapshotRepo, msg amqp.Delivery) error {
	var evt event.UserEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		return fmt.Errorf("unmarshal user event: %w", err)
	}

	switch evt.Type {
	case event.UserUpserted:
		if evt.Payload.UserID == 0 {
			return fmt.Errorf("invalid user_id")
		}
		return userRepo.Upsert(model.UserSnapshot{
			UserID:    evt.Payload.UserID,
			Username:  evt.Payload.Username,
			Role:      int(evt.Payload.Role),
			AvatarURL: evt.Payload.AvatarURL,
		})
	case event.UserDeleted:
		if evt.Payload.UserID == 0 {
			return fmt.Errorf("invalid user_id")
		}
		// Keep a tombstone record so downstream permission checks and UI rendering
		// can still work without calling user-service.
		return userRepo.Upsert(model.UserSnapshot{
			UserID:    evt.Payload.UserID,
			Username:  "用户已注销",
			Role:      0,
			AvatarURL: "",
		})
	default:
		// Ignore unknown event types to keep consumer forward-compatible.
		return nil
	}
}

func startUserWorker(ctx context.Context, mq *client.RabbitMQClient, userRepo repository.UserSnapshotRepo) {
	queueName := "forum.user.events"
	bindingKey := "user.*"
	dlqRoutingKey := "forum.user.events.dlq"

RECONNECT:
	if ctx.Err() != nil {
		return
	}

	ch, err := mq.Conn.Channel()
	if err != nil {
		utils.Logger().WithField("service", "forum-service").WithError(err).Error("[UserConsumer] create channel failed")
		time.Sleep(time.Second)
		goto RECONNECT
	}

	// Declare queue (durable) bound to the shared topic exchange, with its own DLQ.
	if err := client.DeclareQueueWithDLQ(ch, queueName, bindingKey, 0, dlqRoutingKey); err != nil {
		_ = ch.Close()
		time.Sleep(time.Second)
		goto RECONNECT
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		false, // autoAck
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		utils.Logger().WithField("service", "forum-service").WithError(err).Error("[UserConsumer] consume failed")
		_ = ch.Close()
		time.Sleep(time.Second)
		goto RECONNECT
	}

	notify := ch.NotifyClose(make(chan *amqp.Error, 1))
	utils.Logger().WithField("service", "forum-service").WithField("queue", queueName).WithField("binding_key", bindingKey).Info("[UserConsumer] consuming")

	for {
		select {
		case <-ctx.Done():
			_ = ch.Close()
			return
		case err := <-notify:
			utils.Logger().WithField("service", "forum-service").WithError(err).Warn("[UserConsumer] channel closed, reconnecting")
			_ = ch.Close()
			time.Sleep(time.Second)
			goto RECONNECT
		case msg, ok := <-msgs:
			if !ok {
				utils.Logger().WithField("service", "forum-service").Warn("[UserConsumer] msgs closed, reconnecting")
				_ = ch.Close()
				time.Sleep(time.Second)
				goto RECONNECT
			}

			if err := handleUserEvent(userRepo, msg); err != nil {
				utils.Logger().WithField("service", "forum-service").WithField("routing_key", msg.RoutingKey).WithError(err).Error("[UserConsumer] handle message failed, sending to DLQ")
				_ = msg.Nack(false, false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}
