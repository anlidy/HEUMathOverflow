package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 校验登录凭证是否有效
func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("SessionID")
		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing SessionID"})
			return
		}

		userID, err := rdb.Get(context.Background(), "session:"+sessionID).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}

		// 注入用户信息
		c.Set("userID", userID)
		c.Next()
	}
}
