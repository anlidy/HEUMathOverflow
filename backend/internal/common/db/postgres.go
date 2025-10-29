package db

import (
	"MathOverflow/internal/common/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitPostgres 初始化 PostgreSQL 数据库
func InitPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	var db *gorm.DB
	var err error

	maxRetries := 5                  // 最多重试5次
	retryInterval := 3 * time.Second // 每次间隔3秒

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("PostgreSQL connected:", cfg.DBName)
			return db, nil
		}

		log.Printf("连接 PostgreSQL 失败 (第 %d/%d 次): %v", i, maxRetries, err)

		if i < maxRetries {
			time.Sleep(retryInterval)
			log.Println("正在重试连接 PostgreSQL...")
		}
	}

	return nil, fmt.Errorf("failed to connect postgres after %d retries: %w", maxRetries, err)
}
