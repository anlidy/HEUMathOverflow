package db

import (
	"MathOverflow/internal/common/config"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongo(cfg config.MongoConfig) (*mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(cfg.URI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect mongo: %w", err)
	}

	mongo := client.Database(cfg.Database)
	fmt.Println("MongoDB connected:", cfg.Database)
	return mongo, nil
}
