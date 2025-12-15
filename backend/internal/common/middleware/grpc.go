package middleware

import (
	"MathOverflow/internal/common/utils"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

