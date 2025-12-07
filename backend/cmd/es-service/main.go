package main

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/es-service/consumer"
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.LoadConfig("es.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	rabbit, err := client.InitRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("init rabbitmq failed: %v", err)
	}
	defer rabbit.Close()

	es, err := client.InitESClient(cfg)
	if err != nil {
		log.Fatalf("init es failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	// post消费者
	eg.Go(func() error {
		workerCount := 10                          // 消费者数量
		postHandleFunc := consumer.HandlePostEvent // 消息处理函数
		return consumer.StartPostConsumer(ctx, rabbit, es, workerCount, postHandleFunc)
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("es-service exit with error: %v", err)
	}
}
