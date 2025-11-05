package main

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/user-service/controller"
	"MathOverflow/internal/user-service/model"
	"MathOverflow/internal/user-service/repository"
	"MathOverflow/internal/user-service/router"
	"MathOverflow/internal/user-service/service"
	pb "MathOverflow/proto/user"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	dir, _ := os.Getwd()
	fmt.Println("当前工作目录:", dir)
	// 加载配置

	cfg, err := config.LoadConfig("user.yaml")
	if err != nil {
		panic(err)
	}
	// 初始化pgsql数据库
	pg, err := client.InitPostgres(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	// 初始化redis数据库
	rdb, err := client.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}

	// 初始化minio数据库
	fmt.Println(cfg.Minio)
	minio, err := client.InitMinIO(cfg.Minio)
	if err != nil {
		panic(err)
	}

	// 初始化依赖
	userRepo := repository.NewUserRepository(pg)
	sessionRepo := repository.NewSessionRepository(rdb)
	fileRepo := repository.NewFileRepository(minio)

	userService := service.NewUserService(cfg, userRepo, sessionRepo, fileRepo)
	userController := controller.NewUserController(userService)
	userServer := controller.NewUserServer(userService)

	// 并发启动gin和gRPC
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r := router.SetupRouter(rdb, userController)
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
			Handler: r, // Gin 实现了 http.Handler 接口
		}
		go func() {
			<-ctx.Done()
			srv.Shutdown(context.Background())
			log.Println("User Service Exited.")
		}()
		log.Println("User Service is running...")
		return srv.ListenAndServe()
	})

	eg.Go(func() error {
		lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.ExposePort))
		grpcServer := grpc.NewServer()
		pb.RegisterUserServiceServer(grpcServer, userServer)
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
			log.Println("User GRPC Exited.")
		}()
		log.Println("User GRPC is running...")
		return grpcServer.Serve(lis)
	})

	if err := eg.Wait(); err != nil {
		log.Println("User服务异常退出:", err)
	}
}
