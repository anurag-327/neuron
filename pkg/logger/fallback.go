package logger

import (
	"context"
	"fmt"
	"time"
)

// fallbackLogger is a minimal logger used before SetGlobalLogger is called.
// It prevents crashes and prints to stdout.
type fallbackLogger struct{}

func NewFallbackLogger() Logger {
	return &fallbackLogger{}
}

func (f *fallbackLogger) Log(ctx context.Context, entry LogEntry) error {
	fmt.Printf("[FALLBACK] [%s] %s\n", entry.Level, entry.Message)
	return nil
}
func (f *fallbackLogger) Info(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return f.Log(ctx, LogEntry{Level: LevelInfo, Message: msg})
}
func (f *fallbackLogger) Error(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return f.Log(ctx, LogEntry{Level: LevelError, Message: msg})
}
func (f *fallbackLogger) Warn(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return f.Log(ctx, LogEntry{Level: LevelWarn, Message: msg})
}
func (f *fallbackLogger) Debug(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return f.Log(ctx, LogEntry{Level: LevelDebug, Message: msg})
}
func (f *fallbackLogger) Close() error { return nil }
