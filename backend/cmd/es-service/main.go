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
		return consumer.StartPostConsumer(rabbit, es, 10)
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("es-service exit with error: %v", err)
	}
}
