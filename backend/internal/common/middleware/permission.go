package middleware

import (
	"MathOverflow/internal/user-service/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 验证是否有权限访问
func PermissionMiddleware(rdb *redis.Client, allowRole model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := model.Role(c.GetInt("role"))
		if !model.IsAllowedRole(allowRole, role) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "没有权限进行此操作", "code": http.StatusUnauthorized})
			return
		}
		c.Next()
	}
}
