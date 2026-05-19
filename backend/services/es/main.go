package main

import (
	"MathOverflow/common/client"
	"MathOverflow/common/config"
	"MathOverflow/common/utils"
	"MathOverflow/services/es/internal/consumer"
	"context"
	"errors"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
)

func main() {
	utils.InitLogger("es-service")

	cfg, err := config.LoadConfig("config/es.yaml")
	if err != nil {
		utils.Logger().WithError(err).Fatal("load config failed")
	}

	rabbit, err := client.InitRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		utils.Logger().WithError(err).Fatal("init rabbitmq failed")
	}
	defer rabbit.Close()

	es, err := client.InitESClient(cfg)
	if err != nil {
		utils.Logger().WithError(err).Fatal("init es client failed")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	// post消费者
	eg.Go(func() error {
		postHandleFunc := consumer.HandlePostEvent // 消息处理函数
		return consumer.StartPostConsumer(ctx, rabbit, es, postHandleFunc)
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		utils.Logger().WithError(err).Fatal("es-service exit with error")
	}

}
