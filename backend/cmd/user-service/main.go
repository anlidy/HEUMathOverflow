package main

import (
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/db"
	"MathOverflow/internal/user-service/controller"
	"MathOverflow/internal/user-service/model"
	"MathOverflow/internal/user-service/repository"
	"MathOverflow/internal/user-service/router"
	"MathOverflow/internal/user-service/service"
	"fmt"
	"log"
	"os"
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
	pg, err := db.InitPostgres(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.User{}); err != nil {
		panic(err)
	}

	// 初始化redis数据库
	rdb, err := db.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}

	// 初始化minio数据库
	fmt.Println(cfg.Minio)
	minio, err := db.InitMinIO(cfg.Minio)
	if err != nil {
		panic(err)
	}

	// 初始化依赖
	userRepo := repository.NewUserRepository(pg)
	sessionRepo := repository.NewSessionRepository(rdb)
	fileRepo := repository.NewFileRepository(minio)

	userService := service.NewUserService(cfg, userRepo, sessionRepo, fileRepo)
	userController := controller.NewUserController(userService)
	r := router.SetupRouter(rdb, userController)
	log.Println("User Service is running...")
	r.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
