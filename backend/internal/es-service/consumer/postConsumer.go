package consumer

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/event"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

// post事件消费者
func StartPostConsumer(ctx context.Context, mq *client.RabbitMQClient, es *client.ESClient, workerCount int,
	handleFunc func(ch *amqp.Channel, es *client.ESClient, msgChan amqp.Delivery) error) error {

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
		go StartWorker(mq, es, workers[i], handleFunc)
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
			if postID <= 0 {
				log.Printf("mq dispatch: invalid postID=%d, drop message: %s", postID, msg.Body)
				msg.Ack(false) // 丢弃
				continue
			}
			shard := int(postID % int64(workerCount))
			workers[shard] <- msg
		}
	}()

	go func() {
		<-ctx.Done()
		log.Println("post consumer shutting down...")
		// 各 shard 关闭
		for _, ch := range workers {
			close(ch)
		}
	}()
	return nil
}

// 处理Post类型的事件
func HandlePostEvent(ch *amqp.Channel, es *client.ESClient, msg amqp.Delivery) error {
	// 解析post事件结构
	var evt event.ForumPostEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		log.Printf("[es-service]: unmarshal event failed: %v\n", err)
		return err
	}
	// 提取文档
	docID := strconv.FormatInt(evt.Payload.ID, 10)

	// 根据事件类型处理
	switch evt.Type {
	case event.ForumPostCreated, event.ForumPostUpdated:
		return handleUpsertPost(es, docID, evt.Payload)

	case event.ForumPostDeleted:
		return handleDeletePost(es, docID)

	default:
		log.Printf("[es-service]: unknown event type: %s\n", evt.Type)
		return nil // 忽略未知事件并 ack
	}
}

// 创建/更新Post的操作
func handleUpsertPost(es *client.ESClient, docID string, payload event.ForumPostPayload) error {
	// 序列化 ES 文档
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[es-service]: marshal payload failed: %v\n", err)
		return err
	}

	// 全量覆盖写入 ES
	resp, err := es.Client.Index(
		es.Index,
		bytes.NewReader(body),
		es.Client.Index.WithDocumentID(docID),
		es.Client.Index.WithRefresh("false"), // 性能更好
	)
	if err != nil {
		log.Printf("[es-service]: index post error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	// 处理非成功状态码
	if resp.IsError() {
		log.Printf("[es-service]: index response error: [%s] %s\n", resp.Status(), resp.String())
		return fmt.Errorf("index error: %s", resp.Status())
	}

	log.Printf("[es-service]: upserted document id=%s\n", docID)
	return nil
}

// 删除Post的操作
func handleDeletePost(es *client.ESClient, docID string) error {
	resp, err := es.Client.Delete(
		es.Index,
		docID,
		es.Client.Delete.WithRefresh("false"),
	)
	if err != nil {
		log.Printf("[es-service]: delete error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	// 允许 404（文档可能早已被删）
	if resp.IsError() && resp.StatusCode != 404 {
		log.Printf("[es-service]: delete response error: [%s] %s\n", resp.Status(), resp.String())
		return fmt.Errorf("delete error: %s", resp.Status())
	}

	log.Printf("[es-service]: deleted document id=%s\n", docID)
	return nil
}
