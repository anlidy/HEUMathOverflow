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
	r.Use(
		middleware.RequestIDMiddleware(),
		middleware.MetricsMiddleware("user-service"),
		middleware.LoggingMiddleware(),
	)

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	user := r.Group("/user")

	// 不需要鉴权
	user.POST("/login", uc.UserLogin)
	user.POST("/register", uc.UserRegister)
	user.GET("/avatar/:filename", uc.UserDownloadAvatar)

	// 需要鉴权
	authUser := user.Group("")
	authUser.POST("/logout", uc.UserLogout)
	authUser.POST("/avatar", uc.UserUploadAvatar)
	authUser.PATCH("/profile", uc.UserUploadProfie)
	authUser.PATCH("/password", uc.UserUpdatePassword)
	authUser.DELETE("/account", uc.UserDeleteAccount)

	return r
}
