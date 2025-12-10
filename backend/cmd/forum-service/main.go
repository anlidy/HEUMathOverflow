package main

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/forum-service/controller"
	"MathOverflow/internal/forum-service/model"
	"MathOverflow/internal/forum-service/repository"
	"MathOverflow/internal/forum-service/router"
	"MathOverflow/internal/forum-service/service"
	"MathOverflow/internal/forum-service/worker"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	dir, _ := os.Getwd()
	fmt.Println("当前工作目录:", dir)
	// 加载配置

	cfg, err := config.LoadConfig("forum.yaml")
	if err != nil {
		panic(err)
	}

	// 初始化RabbitMQ
	rabbit, err := client.InitRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		panic(err)
	}
	defer rabbit.Close()

	// 初始化redis数据库
	rdb, err := client.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}

	// 初始化pgsql数据库
	pg, err := client.InitPostgres(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	// 迁移表
	if err := pg.AutoMigrate(&model.Post{}, &model.PostLike{}, &model.PostStar{}); err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.Reply{}, &model.ReplyLike{}); err != nil {
		panic(err)
	}
	// 添加triggers
	InitTriggers(pg)

	// 初始化minio数据库
	mc, err := client.InitMinIO(cfg.Minio)
	if err != nil {
		panic(err)
	}

	es, err := client.InitESClient(cfg)
	if err != nil {
		log.Fatalf("init es failed: %v", err)
	}

	// 初始化服务
	fileRepo := repository.NewFileRepository(mc)
	postRepo := repository.NewPostRepository(pg, rdb)
	replyRepo := repository.NewReplyRepository(pg, rdb)

	forumServ := service.NewForumService(cfg, fileRepo)
	postServ := service.NewPostService(cfg, rabbit, postRepo, fileRepo)
	replyServ := service.NewReplyService(cfg, replyRepo, fileRepo)
	searchServ := service.NewSearchService(cfg, es, postRepo)

	forumContrller := controller.NewForumController(forumServ)
	postController := controller.NewPostController(postServ)
	replyController := controller.NewReplyController(replyServ)
	searchController := controller.NewSearchController(searchServ)

	// 初始化worker
	postWorker := worker.NewPostWorker(pg, rdb, rabbit)

	// 并发启动
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		interval := 5 * time.Second
		var count int64 = 500
		go postWorker.WriteBackPostWorker(ctx, interval, count)
		return nil
	})

	eg.Go(func() error {
		r := router.SetupRouter(rdb, forumContrller, postController, replyController, searchController)
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
			Handler: r, // Gin 实现了 http.Handler 接口
		}
		go func() {
			<-ctx.Done()
			srv.Shutdown(context.Background())
			log.Println("Forum Service Exited.")
		}()
		log.Println("Forum Service is running...")
		return srv.ListenAndServe()
	})

	if err := eg.Wait(); err != nil {
		log.Println("Forum服务异常退出:", err)
	}

}
