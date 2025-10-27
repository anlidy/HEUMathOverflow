package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 验证是否有权限访问
func PermissionMiddleware(pg *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		//userID := c.GetInt64("userID")
	}
}
