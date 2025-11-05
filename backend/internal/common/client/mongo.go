package client

import (
	"MathOverflow/internal/common/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongo(cfg config.MongoConfig) (*mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)

	var client *mongo.Client
	var err error

	maxRetries := 5                  // 最多重试5次
	retryInterval := 3 * time.Second // 每次间隔3秒

	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err = mongo.Connect(ctx, clientOpts)
		if err != nil {
			log.Printf("连接 MongoDB 失败 (第 %d/%d 次): %v", i, maxRetries, err)
		} else {
			// 尝试 ping 检查连接是否真正可用
			if pingErr := client.Ping(ctx, nil); pingErr != nil {
				log.Printf("Ping MongoDB 失败 (第 %d/%d 次): %v", i, maxRetries, pingErr)
				err = pingErr
			} else {
				fmt.Println("MongoDB connected:", cfg.Database)
				return client.Database(cfg.Database), nil
			}
		}

		if i < maxRetries {
			time.Sleep(retryInterval)
			log.Println("正在重试连接 MongoDB...")
		}
	}

	return nil, fmt.Errorf("failed to connect mongo after %d retries: %w", maxRetries, err)
}
