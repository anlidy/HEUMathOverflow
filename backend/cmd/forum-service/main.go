package main

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/forum-service/controller"
	"MathOverflow/internal/forum-service/model"
	"MathOverflow/internal/forum-service/repository"
	"MathOverflow/internal/forum-service/router"
	"MathOverflow/internal/forum-service/service"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

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
	// 迁移post表
	if err := pg.AutoMigrate(&model.Post{}); err != nil {
		panic(err)
	}
	// 迁移reply表
	if err := pg.AutoMigrate(&model.Reply{}); err != nil {
		panic(err)
	}

	// 初始化mongoDB数据库
	mdb, err := client.InitMongo(cfg.MongoDB)
	if err != nil {
		panic(err)
	}

	// 初始化minio数据库
	mc, err := client.InitMinIO(cfg.Minio)
	if err != nil {
		panic(err)
	}

	// 初始化服务
	fileRepo := repository.NewFileRepository(mc)
	postRepo := repository.NewPostRepository(pg, mdb)
	replyRepo := repository.NewReplyRepository(pg, mdb)

	forumServ := service.NewForumService(cfg, fileRepo)
	postServ := service.NewPostService(cfg, postRepo, fileRepo)
	replyServ := service.NewReplyService(cfg, replyRepo, fileRepo)

	forumContrller := controller.NewForumController(forumServ)
	postController := controller.NewPostController(postServ)
	replyController := controller.NewReplyController(replyServ)

	// 并发启动gin和gRPC
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r := router.SetupRouter(rdb, forumContrller, postController, replyController)
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
