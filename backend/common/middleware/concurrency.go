package middleware

import (
	"MathOverflow/common/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GlobalConcurrencyMiddleware caps the number of in-flight requests handled by the process.
func GlobalConcurrencyMiddleware(maxInflight int) gin.HandlerFunc {
	if maxInflight <= 0 { // 表示不设并发限制
		return func(c *gin.Context) {
			c.Next()
		}
	}

	semaphore := make(chan struct{}, maxInflight)

	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
			defer func() {
				<-semaphore
			}()
			c.Next()
		default:
			utils.WithContext(c.Request.Context()).WithFields(map[string]any{
				"max_inflight": maxInflight,
				"client_ip":    c.ClientIP(),
				"path":         c.Request.URL.Path,
				"method":       c.Request.Method,
			}).Warn("gateway concurrency limit exceeded")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"message": "当前请求过多，请稍后再试",
				"code":    http.StatusServiceUnavailable,
			})
		}
	}
}
