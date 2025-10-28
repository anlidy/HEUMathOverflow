package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 校验登录凭证是否有效
// AuthMiddleware 校验登录凭证是否有效（从 Cookie 中读取 SessionID）
func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Cookie 中读取 sessionID
		sessionID, err := c.Cookie("session-id")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing session-id"})
			return
		}

		// 去 Redis 校验 session 是否存在
		userIDStr, err := rdb.HGet(context.Background(), "session:"+sessionID, "userID").Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			return
		}

		// 注入用户信息到上下文
		userID, _ := strconv.ParseInt(userIDStr, 10, 64) // 转成 int64
		c.Set("userID", userID)

		c.Next()
	}
}
