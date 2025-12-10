package consumer

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/event"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PostHandleFunc func(ch *amqp.Channel, es *client.ESClient, msg amqp.Delivery) error

// post事件消费者
// StartPostConsumers 启动 N 个分片 worker，每个 worker 消费一个独立队列
// 分片规则：生产者按 PostID % workerCount 选择 routingKey -> shard 队列
func StartPostConsumer(ctx context.Context, mq *client.RabbitMQClient, es *client.ESClient,
	handleFunc PostHandleFunc) error {

	if mq.PostWorkerCount <= 0 {
		return fmt.Errorf("workerCount must > 0")
	}

	// 每个 shard 启一个 worker
	for shardID := 0; shardID < int(mq.PostWorkerCount); shardID++ {
		go StartWorker(ctx, mq, es, shardID, handleFunc)
	}

	log.Printf("[PostConsumer] %d shard workers started.\n", mq.PostWorkerCount)

	// 等待上游关闭
	<-ctx.Done()
	log.Println("[PostConsumer] context canceled, waiting workers to exit...")
	return nil
}

// 获取帖子状态
func getCurrentPostStat(es *client.ESClient, index, docID string) (*common.PostStat, error) {
	res, err := es.Client.Get(index, docID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("get doc failed: %s", res.String())
	}

	var doc struct {
		Source common.PostStat `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &doc.Source, nil
}

// 处理Post类型的事件
func HandlePostEvent(ch *amqp.Channel, es *client.ESClient, msg amqp.Delivery) error {
	// 解析post事件结构
	var evt event.ForumPostEvent
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		log.Printf("[es-service]: unmarshal event failed: %v\n", err)
		return err
	}
	log.Println(evt)
	// 提取文档
	docID := strconv.FormatInt(evt.Payload.PostID, 10)

	// 根据事件类型处理
	switch evt.Type {
	case event.ForumPostCreated:
		return handleCreatePost(es, docID, evt.Payload)
	case event.ForumPostUpdated:
		return handleUpdatePost(es, docID, evt.Payload)
	case event.ForumPostDeleted:
		return handleDeletePost(es, docID)
	case event.ForumPostStatUpdated:
		return handlePostStatUpdated(es, docID, evt.Payload)
	default:
		log.Printf("[es-service]: unknown event type: %s\n", evt.Type)
		return nil // 忽略未知事件并 ack
	}
}

// 创建Post的操作
func handleCreatePost(es *client.ESClient, docID string, payload event.ForumPostPayload) error {
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

	log.Printf("[es-service]: create document id=%s\n", docID)
	return nil
}

func handleUpdatePost(es *client.ESClient, docID string, payload event.ForumPostPayload) error {

	updateBody := map[string]interface{}{
		"doc": payload, // 用 doc 包裹，ES 才能识别
	}

	body, err := json.Marshal(updateBody)
	if err != nil {
		log.Printf("[es-service]: marshal payload failed: %v\n", err)
		return err
	}

	resp, err := es.Client.Update(
		es.Index,
		docID,
		bytes.NewReader(body),
		es.Client.Update.WithRefresh("false"),
	)
	if err != nil {
		log.Printf("[es-service]: Update post error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		log.Printf("[es-service]: Update response error: [%s] %s\n", resp.Status(), resp.String())
		return fmt.Errorf("update error: %s", resp.Status())
	}

	log.Printf("[es-service]: update document id=%s\n", docID)
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

// 帖子状态更新
func handlePostStatUpdated(es *client.ESClient, docID string, payload event.ForumPostPayload) error {
	// 查询当前状态
	stat, err := getCurrentPostStat(es, es.Index, docID)
	if err != nil {
		return err
	}
	// 更新状态
	stat.Likes += payload.Likes
	stat.Replies += payload.Replies
	stat.Stars += payload.Stars
	stat.Views += payload.Views

	// 计算评分
	favors, totalScore := utils.CalcPostScores(stat)
	stat.Favors = favors
	stat.TotalScore = totalScore

	updateBody := map[string]any{
		"doc": stat, // 用 doc 包裹，ES 才能识别
	}

	// 序列化 ES 文档
	body, err := json.Marshal(updateBody)
	if err != nil {
		log.Printf("[es-service]: marshal payload failed: %v\n", err)
		return err
	}

	res, err := es.Client.Update(
		es.Index, // index
		docID,    // document ID
		bytes.NewReader(body),
		es.Client.Update.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("ES update error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES update failed for post %s", docID)
	}
	return nil
}
