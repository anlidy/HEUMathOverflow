package client

import (
	"MathOverflow/internal/common/config"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Conn *amqp.Connection
}

// InitRabbitMQ 初始化 RabbitMQ 客户端，包含连接与默认通道
func InitRabbitMQ(cfg config.RabbitMQConfig) (*RabbitMQClient, error) {
	uri := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	maxRetries := 5
	retryInterval := 3 * time.Second

	var (
		conn *amqp.Connection
		err  error
	)

	for i := 1; i <= maxRetries; i++ {
		conn, err = amqp.Dial(uri)
		if err == nil {
			fmt.Println("RabbitMQ connected:", uri)
			return &RabbitMQClient{Conn: conn}, nil
		}

		log.Printf("连接 RabbitMQ 失败 (第 %d/%d 次): %v", i, maxRetries, err)
		if i < maxRetries {
			time.Sleep(retryInterval)
			log.Println("正在重试连接 RabbitMQ...")
		}
	}

	return nil, fmt.Errorf("failed to connect rabbitmq after %d retries: %w", maxRetries, err)
}

// Close 关闭频道与连接
func (c *RabbitMQClient) Close() error {
	if c == nil {
		return nil
	}
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

// 生产者-事件发布
func (c *RabbitMQClient) PublishEvent(eventType string, payload any) error {
	// 业务数据序列化成 JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mq payload marshal error: %w", err)
	}
	// 创建临时channel
	ch, err := c.Conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 声明交换机
	// O(1)操作,不会重复或影响性能
	err = ch.ExchangeDeclare("app.events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// 发布消息到交换机
	return ch.Publish("app.events", eventType, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// 消费者-声明队列
func DeclareQueue(ch *amqp.Channel, qname string, key string) error {
	// 声明交换机
	err := ch.ExchangeDeclare("app.events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}
	// 声明queue
	_, err = ch.QueueDeclare(qname, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return ch.QueueBind(qname, key, "app.events", false, nil)
}
