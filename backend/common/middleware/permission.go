package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 验证是否有权限访问
func PermissionMiddleware(rdb *redis.Client, allowRole int) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetInt("role")
		if role < allowRole {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "没有权限进行此操作", "code": http.StatusUnauthorized})
			return
		}
		c.Next()
	}
}
