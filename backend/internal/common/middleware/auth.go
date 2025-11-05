package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 校验登录凭证是否有效
// AuthMiddleware 校验登录凭证是否有效（从 Cookie 中读取 SessionID）
func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// 从 Cookie 中读取 sessionID
		sessionID, err := c.Cookie("session-id")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录再进行此操作", "code": http.StatusUnauthorized})
			return
		}

		// 去 Redis 校验 session 是否存在
		data, err := rdb.HGetAll(ctx, "session:"+sessionID).Result()
		if err != nil || len(data) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "会话已过期,请重新登录", "code": http.StatusUnauthorized})
			return
		}

		// 解析 role
		roleStr, err := rdb.Get(ctx, data["roleKey"]).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "会话已过期,请重新登录", "code": http.StatusUnauthorized})
			return
		}
		role, _ := strconv.ParseInt(roleStr, 10, 0)

		// 解析 userID
		userID, _ := strconv.ParseInt(data["userID"], 10, 64)

		// 注入用户信息到上下文
		c.Set("userID", userID)
		c.Set("role", int(role))
		c.Next()
	}
}
