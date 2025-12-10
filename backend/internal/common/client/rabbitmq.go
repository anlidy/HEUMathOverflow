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
	Conn            *amqp.Connection
	PostWorkerCount int64
}

const (
	exchangeName = "app.events"
	exchangeType = "topic"
)

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
			log.Printf("PostWorker Count: %d\n", cfg.PostWorkerCount)
			return &RabbitMQClient{
				Conn:            conn,
				PostWorkerCount: cfg.PostWorkerCount,
			}, nil
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
func (c *RabbitMQClient) PublishEvent(routeKey string, payload any) error {
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
	err = ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// 发布消息到交换机
	return ch.Publish(exchangeName, routeKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// 消费者-声明队列
func DeclareQueue(ch *amqp.Channel, queueName string, bindingKey string, shardID int) error {
	// QoS：每次只处理一条，处理完再给下一条，避免一个 worker 堆积太多未 ack 消息
	if err := ch.Qos(1, 0, false); err != nil {
		log.Printf("[Shard %d] set QoS failed: %v\n", shardID, err)
		return err
	}

	// 声明 exchange（幂等）
	if err := ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  // durable
		false, // autoDelete
		false,
		false,
		nil,
	); err != nil {
		log.Printf("[Shard %d] declare exchange failed: %v\n", shardID, err)
		return err
	}

	// 声明自己的 shard 队列
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		log.Printf("[Shard %d] declare queue failed: %v\n", shardID, err)
		return err
	}

	// 绑定队列到 exchange 上
	if err := ch.QueueBind(
		q.Name,
		bindingKey, // 只接收自己 shard 的消息
		exchangeName,
		false,
		nil,
	); err != nil {
		log.Printf("[Shard %d] bind queue failed: %v\n", shardID, err)
		return err
	}
	return nil
}
