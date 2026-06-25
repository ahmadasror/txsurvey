// Package logging configures the global slog JSON logger and carries a
// request-id through context for correlated logs.
package logging

import (
	"context"
	"log/slog"
	"os"
)

// Setup initializes the global slog logger with JSON output to stdout.
func Setup() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
}

type ctxKey string

const requestIDKey ctxKey = "request_id"

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext retrieves the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// With returns a logger enriched with request_id when present in context.
func With(ctx context.Context) *slog.Logger {
	logger := slog.Default()
	if id := RequestIDFromContext(ctx); id != "" {
		logger = logger.With("request_id", id)
	}
	return logger
}
