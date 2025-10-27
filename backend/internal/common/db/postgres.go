package db

import (
	"MathOverflow/internal/common/config"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitPostgres 初始化 PostgreSQL 数据库
func InitPostgres(cfg config.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres: %w", err)
	}

	fmt.Println("PostgreSQL connected:", cfg.DBName)
	return db, nil
}
