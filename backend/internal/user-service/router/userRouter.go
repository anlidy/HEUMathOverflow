package router

import (
	"MathOverflow/internal/common/middleware"
	"MathOverflow/internal/user-service/controller"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client, uc controller.UserController) *gin.Engine {
	r := gin.Default()
<<<<<<< HEAD
	r.Use(middleware.CorsMiddleware([]string{"http://localhost:5173", "https://math-overflow.edu"}))
=======
	r.Use(
		middleware.RequestIDMiddleware(),
		middleware.MetricsMiddleware("user-service"),
		middleware.LoggingMiddleware(),
		middleware.CorsMiddleware([]string{"http://localhost:3000", "https://math-overflow.edu"}),
	)

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

>>>>>>> go-dev
	api := r.Group("/api")
	v1 := api.Group("/v1")

	user := v1.Group("/user")

	// 不需要鉴权
	user.POST("/login", uc.UserLogin)
	user.POST("/register", uc.UserRegister)
	user.GET("/avatar/:filename", uc.UserDownloadAvatar)

	// 需要鉴权
	authUser := user.Group("")
	authUser.Use(middleware.AuthMiddleware(rdb))
	authUser.POST("/avatar", uc.UserUploadAvatar)
	authUser.PATCH("/profile", uc.UserUploadProfie)
	authUser.PATCH("/password", uc.UserUpdatePassword)

	return r
}
