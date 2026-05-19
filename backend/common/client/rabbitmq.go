package client

import (
	"MathOverflow/common/config"
	"MathOverflow/common/utils"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Conn            *amqp.Connection
	PostWorkerCount int64
	RagWorkerCount  int64
	breaker         *utils.CircuitBreaker
}

const (
	exchangeName    = "app.events"
	exchangeType    = "topic"
	postDLQExchange = "app.events.dlx"
	postDLQName     = "es.post.events.dlq"
	userDLQName     = "forum.user.events.dlq"
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
			utils.Logger().WithField("post_worker_count", cfg.PostWorkerCount).
				WithField("rag_worker_count", cfg.RagWorkerCount).
				Info("RabbitMQ connected")
			mq := &RabbitMQClient{
				Conn:            conn,
				PostWorkerCount: cfg.PostWorkerCount,
				RagWorkerCount:  cfg.RagWorkerCount,
				breaker:         utils.NewCircuitBreaker(5, 5*time.Second),
			}

			// 初始化死信交换机与公共 DLQ（幂等操作）
			ch, err := conn.Channel()
			if err != nil {
				utils.Logger().WithError(err).Error("declare dead-letter exchange/channel failed")
				return mq, nil
			}
			if err := ch.ExchangeDeclare(
				postDLQExchange,
				"direct",
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				utils.Logger().WithError(err).Error("declare dead-letter exchange failed")
				_ = ch.Close()
				return mq, nil
			}
			// 声明一个通用的 DLQ，用于接收消费失败的消息
			if _, err := ch.QueueDeclare(
				postDLQName,
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				utils.Logger().WithError(err).Error("declare dead-letter queue failed")
				_ = ch.Close()
				return mq, nil
			}
			if err := ch.QueueBind(
				postDLQName,
				postDLQName,
				postDLQExchange,
				false,
				nil,
			); err != nil {
				utils.Logger().WithError(err).Error("bind dead-letter queue failed")
				_ = ch.Close()
				return mq, nil
			}

			// 声明一个通用的 User DLQ，用于接收 user 事件消费失败的消息（幂等）
			if _, err := ch.QueueDeclare(
				userDLQName,
				true,
				false,
				false,
				false,
				nil,
			); err != nil {
				utils.Logger().WithError(err).Error("declare user dead-letter queue failed")
				_ = ch.Close()
				return mq, nil
			}
			if err := ch.QueueBind(
				userDLQName,
				userDLQName,
				postDLQExchange,
				false,
				nil,
			); err != nil {
				utils.Logger().WithError(err).Error("bind user dead-letter queue failed")
				_ = ch.Close()
				return mq, nil
			}
			_ = ch.Close()

			return mq, nil
		}

		utils.Logger().WithField("retry", i).WithField("max_retries", maxRetries).WithError(err).Error("连接 RabbitMQ 失败")
		if i < maxRetries {
			time.Sleep(retryInterval)
			utils.Logger().Info("正在重试连接 RabbitMQ...")
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
func (c *RabbitMQClient) PublishEvent(routeKey string, body []byte) error {
	if c == nil {
		return fmt.Errorf("mq client is nil")
	}

	if c.breaker != nil && !c.breaker.Allow() {
		return fmt.Errorf("rabbitmq circuit breaker is open")
	}

	const (
		maxRetries     = 2
		initialBackoff = 100 * time.Millisecond
	)

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 创建临时channel
		ch, err := c.Conn.Channel()
		if err != nil {
			lastErr = err
		} else {
			// 确保关闭
			// 声明交换机（幂等）
			err = ch.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)
			if err == nil {
				// 发布消息到交换机
				err = ch.Publish(exchangeName, routeKey, false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				})
			}
			_ = ch.Close()
			lastErr = err
		}

		if lastErr == nil {
			if c.breaker != nil {
				c.breaker.OnSuccess()
			}
			return nil
		}

		if c.breaker != nil {
			c.breaker.OnFailure()
		}

		if attempt < maxRetries {
			time.Sleep(initialBackoff * time.Duration(attempt+1))
			continue
		}
	}

	return lastErr
}

// 消费者-声明队列（支持指定 DLQ routing key）
func DeclareQueueWithDLQ(ch *amqp.Channel, queueName string, bindingKey string, shardID int, dlqRoutingKey string) error {
	// QoS：每次只处理一条，处理完再给下一条，避免一个 worker 堆积太多未 ack 消息
	if err := ch.Qos(1, 0, false); err != nil {
		utils.Logger().WithField("shard", shardID).WithError(err).Error("set QoS failed")
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
		utils.Logger().WithField("shard", shardID).WithError(err).Error("declare exchange failed")
		return err
	}

	// 声明自己的 shard 队列
	// 启用post死信队列
	args := amqp.Table{
		"x-dead-letter-exchange":    postDLQExchange,
		"x-dead-letter-routing-key": dlqRoutingKey,
	}
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		args,
	)
	if err != nil {
		utils.Logger().WithField("shard", shardID).WithError(err).Error("declare queue failed")
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
		utils.Logger().WithField("shard", shardID).WithError(err).Error("bind queue failed")
		return err
	}
	return nil
}

// DeclareQueue 声明 shard 队列（默认死信进入 post DLQ）
func DeclareQueue(ch *amqp.Channel, queueName string, bindingKey string, shardID int) error {
	return DeclareQueueWithDLQ(ch, queueName, bindingKey, shardID, postDLQName)
}
