package consumer

import (
	"MathOverflow/internal/common/client"
	"context"
	"fmt"
	"log"
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
		log.Printf("[Shard %d] create channel failed: %v\n", shardID, err)
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
		log.Printf("[Shard %d] consume failed: %v\n", shardID, err)
		_ = ch.Close()
		time.Sleep(time.Second)
		goto RECONNECT
	}
	// 监听 channel 关闭
	notify := ch.NotifyClose(make(chan *amqp.Error, 1))
	log.Printf("[Shard %d] worker started, queue=%s, bindingKey=%s\n", shardID, queueName, bindingKey)

	// 持续监听事件
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Shard %d] context done, closing channel...\n", shardID)
			_ = ch.Close()
			return

		case err := <-notify:
			// channel 被动关闭（例如 unknown delivery tag、网络异常等）
			log.Printf("[Shard %d] channel closed: %v, reconnecting...\n", shardID, err)
			_ = ch.Close()
			time.Sleep(time.Second)
			goto RECONNECT

		case msg, ok := <-msgs:
			if !ok {
				// msgs 关闭了，一般是 channel 关闭了
				log.Printf("[Shard %d] msgs channel closed, reconnecting...\n", shardID)
				_ = ch.Close()
				time.Sleep(time.Second)
				goto RECONNECT
			}

			// 开始处理post事件
			if err := handleFunc(ch, es, msg); err != nil {
				log.Printf("[Shard %d] handle message error: %v, nack & requeue\n", shardID, err)
				time.Sleep(time.Second)
				_ = msg.Nack(false, true)
			} else {
				log.Printf("[Shard %d] message handled ok, ack. deliveryTag=%d\n", shardID, msg.DeliveryTag)
				_ = msg.Ack(false)
			}
		}
	}
}
