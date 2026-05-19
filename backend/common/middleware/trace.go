package middleware

import (
	"MathOverflow/common/utils"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const ginTraceIDKey = "trace_id"

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
		c.Set(ginTraceIDKey, traceID)
		c.Writer.Header().Set("X-Request-Id", traceID)

		// Propagate traceID into the request context for non-HTTP layers.
		ctxWithTrace := utils.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctxWithTrace)

		c.Next()
	}
}

// UnaryServerTraceInterceptor extracts or assigns a trace ID for each
// incoming gRPC request and injects it into the context so that
// downstream code can log with trace-aware utilities.
func UnaryServerTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var traceID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-request-id"); len(vals) > 0 {
				traceID = vals[0]
			}
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx = utils.ContextWithTraceID(ctx, traceID)
		return handler(ctx, req)
	}
}
