package log

import (
	"context"
)

type loggerKey struct{}

// WithLoger adds a logger to a given context.
func WithLogger(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// GetLogger returns the logger from a given context.
//
// If no logger is present on the context a shardd discard logger is returned
// which will ignore logging calls (but is still a valid reference!)
func GetLogger(ctx context.Context) *Logger {
	if value := ctx.Value(loggerKey{}); value != nil {
		if typed, ok := value.(*Logger); ok {
			return typed
		}
	}
	return Discard()
}
