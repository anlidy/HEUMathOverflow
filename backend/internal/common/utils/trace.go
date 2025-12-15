package utils

import "context"

// traceIDKey is an unexported type for keys defined in this package.
// This prevents collisions with context keys defined in other packages.
type traceIDKey struct{}

var traceKey = traceIDKey{}

// ContextWithTraceID returns a new context with the given traceID attached.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceKey, traceID)
}

// TraceIDFromContext extracts the traceID from the given context if present.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(traceKey)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

