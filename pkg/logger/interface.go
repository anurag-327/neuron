package logger

import (
	"context"
	"time"
)

// Logger is the main interface for pushing logs.
// Services should depend on this interface, not concrete structs.
type Logger interface {
	// Log sends a generic log entry
	Log(ctx context.Context, entry LogEntry) error

	// Helper functions for common log levels
	Info(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error
	Error(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error
	Warn(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error
	Debug(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error

	// Close cleans up underlying connections
	Close() error
}

// Consumer is the interface for workers consuming logs.
type Consumer interface {
	// FetchBatch waits for logs and returns up to batchSize items.
	FetchBatch(ctx context.Context, batchSize int, timeout time.Duration) ([]LogEntry, error)
	// Close cleans up underlying connections
	Close() error
}
