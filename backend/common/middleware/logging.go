package middleware

import (
	"MathOverflow/common/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware writes a structured access log per HTTP request using logrus.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		traceID, _ := c.Get(ginTraceIDKey)

		entry := utils.Logger().WithFields(logrus.Fields{
			"service":    utils.ServiceName(),
			"trace_id":   traceID,
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"query":      rawQuery,
		})

		if len(c.Errors) > 0 {
			entry.WithField("errors", c.Errors.String()).Error("request completed with errors")
			return
		}

		switch {
		case statusCode >= 500:
			entry.Error("server error")
		case statusCode >= 400:
			entry.Warn("client error")
		default:
			entry.Info("request completed")
		}
	}
}
