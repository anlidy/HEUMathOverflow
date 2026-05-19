package router

import (
	"MathOverflow/common/middleware"
	handler "MathOverflow/services/user/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client, uh handler.UserHandler) *gin.Engine {
	r := gin.Default()
	r.Use(
		middleware.RequestIDMiddleware(),
		middleware.LoggingMiddleware(),
	)

	user := r.Group("/user")

	// 不需要鉴权
	user.POST("/login", uh.UserLogin)
	user.POST("/register", uh.UserRegister)
	user.GET("/avatar/:filename", uh.UserDownloadAvatar)

	// 需要鉴权
	authUser := user.Group("")
	authUser.POST("/logout", uh.UserLogout)
	authUser.POST("/avatar", uh.UserUploadAvatar)
	authUser.PATCH("/profile", uh.UserUploadProfie)
	authUser.PATCH("/password", uh.UserUpdatePassword)
	authUser.DELETE("/account", uh.UserDeleteAccount)

	return r
}
