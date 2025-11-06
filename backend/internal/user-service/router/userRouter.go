package router

import (
	"MathOverflow/internal/common/middleware"
	"MathOverflow/internal/user-service/controller"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client, uc controller.UserController) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CorsMiddleware([]string{"http://localhost:5173", "https://math-overflow.edu"}))
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
	authUser.PATCH("/profile")
	authUser.PATCH("/password")

	return r
}
