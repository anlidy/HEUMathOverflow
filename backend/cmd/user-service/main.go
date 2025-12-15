package main

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/middleware"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/user-service/controller"
	"MathOverflow/internal/user-service/model"
	"MathOverflow/internal/user-service/repository"
	"MathOverflow/internal/user-service/router"
	"MathOverflow/internal/user-service/service"
	pb "MathOverflow/proto/user"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	utils.InitLogger("user-service")

	dir, _ := os.Getwd()
	utils.Logger().WithField("working_dir", dir).Info("user-service starting")
	// 加载配置

	cfg, err := config.LoadConfig("user.yaml")
	if err != nil {
		utils.Logger().WithError(err).Fatal("failed to load user.yaml")
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
			utils.Logger().Info("User Service Exited.")
		}()
		utils.Logger().Info("User Service is running...")
		return srv.ListenAndServe()
	})

	eg.Go(func() error {
		lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.ExposePort))
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(middleware.UnaryServerTraceInterceptor()),
		)
		pb.RegisterUserServiceServer(grpcServer, &userServer)
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
			utils.Logger().Info("User GRPC Exited.")
		}()
		utils.Logger().WithField("grpc_port", cfg.GRPC.ExposePort).Info("User GRPC is running...")
		return grpcServer.Serve(lis)
	})

	if err := eg.Wait(); err != nil {
		utils.Logger().WithError(err).Error("User服务异常退出")
	}
}
