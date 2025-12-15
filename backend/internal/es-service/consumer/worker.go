package consumer

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/utils"
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartWorker(
	ctx context.Context,
	mq *client.RabbitMQClient,
	es *client.ESClient,
	shardID int,
	handleFunc PostHandleFunc) {

	queueName := fmt.Sprintf("es.post.events.shard-%d", shardID)
	// routing key 约定：论坛帖子事件
	// 生产者发送时使用：forum.post.*.<shardID>，以保证同一 Post 进同一 shard
	bindingKey := fmt.Sprintf("forum.post.*.%d", shardID)

RECONNECT:
	if ctx.Err() != nil {
		return
	}
	// 建立一个channel
	ch, err := mq.Conn.Channel()
	if err != nil {
		utils.Logger().WithField("service", "es-service").WithField("shard", shardID).WithError(err).Error("create channel failed")
		time.Sleep(time.Second)
		goto RECONNECT
	}
	// 声明队列
	err = client.DeclareQueue(ch, queueName, bindingKey, shardID)
	if err != nil {
		_ = ch.Close()
		goto RECONNECT
	}

	// 开始消费
	msgs, err := ch.Consume(
		queueName,
		"",    // consumer tag
		false, // autoAck = false 手动 ack
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		utils.Logger().WithField("service", "es-service").WithField("shard", shardID).WithError(err).Error("consume failed")
		_ = ch.Close()
		time.Sleep(time.Second)
		goto RECONNECT
	}
	// 监听 channel 关闭
	notify := ch.NotifyClose(make(chan *amqp.Error, 1))
	utils.Logger().WithField("service", "es-service").WithField("shard", shardID).WithField("queue", queueName).WithField("binding_key", bindingKey).Info("worker started")

	// 持续监听事件
	for {
		select {
		case <-ctx.Done():
			utils.Logger().WithField("service", "es-service").WithField("shard", shardID).Info("context done, closing channel")
			_ = ch.Close()
			return

		case err := <-notify:
			// channel 被动关闭（例如 unknown delivery tag、网络异常等）
			utils.Logger().WithField("service", "es-service").WithField("shard", shardID).WithError(err).Warn("channel closed, reconnecting")
			_ = ch.Close()
			time.Sleep(time.Second)
			goto RECONNECT

		case msg, ok := <-msgs:
			if !ok {
				// msgs 关闭了，一般是 channel 关闭了
				utils.Logger().WithField("service", "es-service").WithField("shard", shardID).Warn("msgs channel closed, reconnecting")
				_ = ch.Close()
				time.Sleep(time.Second)
				goto RECONNECT
			}

			// 开始处理post事件
			if err := handleFunc(ch, es, msg); err != nil {
				utils.Logger().
					WithField("service", "es-service").
					WithField("shard", shardID).
					WithField("routing_key", msg.RoutingKey).
					WithError(err).
					Error("handle message error, send to dead-letter queue")
				time.Sleep(time.Second)
				// 不再重回原队列，避免无限重试，由 RabbitMQ 投递到 DLQ
				_ = msg.Nack(false, false)
			} else {
				utils.Logger().WithField("service", "es-service").WithField("shard", shardID).WithField("delivery_tag", msg.DeliveryTag).Info("message handled ok, ack")
				_ = msg.Ack(false)
			}
		}
	}
}
