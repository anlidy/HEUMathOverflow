package client

import (
	"MathOverflow/internal/common/config"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type ESClient struct {
	Client *elasticsearch.Client
	Index  string
}

func InitESClient(cfg config.Config) (*ESClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Elastic.Addresses,
		Username:  cfg.Elastic.Username,
		Password:  cfg.Elastic.Password,
	}

	var esClient *elasticsearch.Client
	var err error
	maxRetries := 5                  // 最多重试5次
	retryInterval := 3 * time.Second // 每次间隔3秒
	for i := 1; i <= maxRetries; i++ {
		esClient, err = elasticsearch.NewClient(esCfg)
		if err == nil {
			fmt.Println("ES connected.")
			return &ESClient{
				Client: esClient,
				Index:  cfg.Elastic.Index,
			}, nil
		}

		log.Printf("连接 ES 失败 (第 %d/%d 次): %v", i, maxRetries, err)

		if i < maxRetries {
			time.Sleep(retryInterval)
			log.Println("正在重试连接 ES...")
		}
	}
	return nil, fmt.Errorf("failed to connect ES after %d retries: %w", maxRetries, err)
}
