package consumer

import (
	"MathOverflow/internal/common/client"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartWorker(
	mq *client.RabbitMQClient,
	es *client.ESClient,
	msgChan chan amqp.Delivery,
	handleFunc func(ch *amqp.Channel, es *client.ESClient, msgChan amqp.Delivery) error) {
	for {
		ch, err := mq.Conn.Channel()
		if err != nil {
			log.Println("worker: failed to create channel:", err)
			time.Sleep(time.Second)
			continue
		}

		// 监听 channel 是否关闭
		notify := ch.NotifyClose(make(chan *amqp.Error))
		log.Println("worker started with new channel")

		for {
			select {
			case msg, ok := <-msgChan:
				if !ok {
					ch.Close()
					return
				}
				// 处理消息
				if err := handleFunc(ch, es, msg); err != nil {
					log.Println("worker process error:", err)
					ch.Nack(msg.DeliveryTag, false, true) // 重新投送
				} else {
					ch.Ack(msg.DeliveryTag, false)
				}

			case err := <-notify:
				log.Println("worker: notify channel closed:", err)
				ch.Close()
				time.Sleep(time.Second)
				goto RESTART
			}
		RESTART:
			continue
		}
	}
}
