package client

import (
	"MathOverflow/common/config"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化 Redis
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println("Redis connected")
	return rdb, nil
}
