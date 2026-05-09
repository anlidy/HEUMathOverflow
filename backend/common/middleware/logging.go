package middleware

import (
	"MathOverflow/common/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const traceIDKey = "trace_id"

// RequestIDMiddleware ensures every request has a trace ID.
// It reads X-Request-Id from the incoming header or generates a new one,
// then propagates it via Gin context, HTTP headers, and request context.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.Request.Header.Get("X-Request-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Expose traceID to downstream handlers and clients.
		c.Set(traceIDKey, traceID)
		c.Writer.Header().Set("X-Request-Id", traceID)

		// Propagate traceID into the request context for non-HTTP layers.
		ctxWithTrace := utils.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctxWithTrace)

		c.Next()
	}
}

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

		traceID, _ := c.Get(traceIDKey)

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
