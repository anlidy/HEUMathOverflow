package consumer

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/event"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// post事件消费者
func StartPostConsumer(mq *client.RabbitMQClient, es *client.ESClient, workerCount int) error {
	ch, err := mq.Conn.Channel()
	if err != nil {
		return err
	}

	// 声明 exchange/queue/bind
	if err := client.DeclareQueue(ch, "es.post.events", "post.*"); err != nil {
		return err
	}

	msgs, err := ch.Consume("es.post.events", "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// 创建 worker pool
	workers := make([]chan amqp.Delivery, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = make(chan amqp.Delivery, 100) // 缓冲区为100
		// 每个worker一个channel,独立ack
		go StartESWorker(mq, es, workers[i])
	}

	// Sharding分片
	go func() {
		for msg := range msgs {
			// 解析event获取id
			var event event.ForumPostEvent
			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				log.Printf("mq dispatch err:%v\n", err)
			}
			// 根据postID进行分派,确保同一个id被同一个消费者消费
			postID := event.Payload.ID
			shard := int(postID % int64(workerCount))
			workers[shard] <- msg
		}
	}()
	return nil
}

func StartESWorker(mq *client.RabbitMQClient, es *client.ESClient, msgChan chan amqp.Delivery) {
	ch, _ := mq.Conn.Channel()
	defer ch.Close()

	for msg := range msgChan {
		// todo
		ch.Ack(msg.DeliveryTag, false)
	}
}
