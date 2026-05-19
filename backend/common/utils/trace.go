package utils

import "context"

// contextTraceIDKey is an unexported type for keys defined in this package.
// This prevents collisions with context keys defined in other packages.
type contextTraceIDKey struct{}

var traceContextKey = contextTraceIDKey{}

// ContextWithTraceID returns a new context with the given traceID attached.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceContextKey, traceID)
}

// TraceIDFromContext extracts the traceID from the given context if present.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(traceContextKey)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
