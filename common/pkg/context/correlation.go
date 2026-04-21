package context

import (
	"context"
	"github.com/google/uuid"
)

type correlationIDKey struct{}

// WithCorrelationID returns a new context with the provided correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// GetCorrelationID extracts the correlation ID from the context.
// If not found, it generates a new one.
func GetCorrelationID(ctx context.Context) string {
	if val, ok := ctx.Value(correlationIDKey{}).(string); ok && val != "" {
		return val
	}
	return uuid.New().String()
}

// EnsureCorrelationID ensures that a context has a correlation ID.
func EnsureCorrelationID(ctx context.Context) context.Context {
	if val, ok := ctx.Value(correlationIDKey{}).(string); ok && val != "" {
		return ctx
	}
	return WithCorrelationID(ctx, uuid.New().String())
}
