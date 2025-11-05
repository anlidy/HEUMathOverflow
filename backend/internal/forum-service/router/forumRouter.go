package router

import (
	"MathOverflow/internal/common/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CorsMiddleware([]string{"http://localhost:3000", "https://math-overflow.edu"}))
	api := r.Group("/api")
	v1 := api.Group("/v1")
	forum := v1.Group("forum")

	// 需要鉴权
	authForum := forum.Group("")
	authForum.Use(middleware.AuthMiddleware(rdb))
	authForum.POST("/posts/upload")
	return r
}
