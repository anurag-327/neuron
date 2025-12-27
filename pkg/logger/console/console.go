package consoleLogger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/pkg/logger"
)

// ConsoleLogger implements the Logger interface but prints to Stdout.
// Useful for local development without Redis.
type ConsoleLogger struct {
	serviceName string
}

// NewConsoleLogger creates a logger that writes to standard output
func NewConsoleLogger(serviceName string) *ConsoleLogger {
	if serviceName == "" {
		serviceName = "console-app"
	}
	return &ConsoleLogger{serviceName: serviceName}
}

// Log prints the entry to stdout in a structured text format
func (c *ConsoleLogger) Log(ctx context.Context, entry logger.LogEntry) error {
	// Defaults
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.Service == "" {
		entry.Service = c.serviceName
	}
	if entry.Level == "" {
		entry.Level = logger.LevelInfo
	}

	// Format: [TIME] [LEVEL] [SERVICE] MESSAGE {METADATA}
	// e.g.: [2024-05-20T10:00:00Z] [INFO] [auth-service] User login {"user_id": 123}

	metaStr := ""
	if len(entry.Metadata) > 0 {
		// Pretty print metadata as JSON string for readability
		if b, err := json.Marshal(entry.Metadata); err == nil {
			metaStr = " " + string(b)
		}
	}

	fmt.Printf("[%s] [%s] [%s] %s%s\n",
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Service,
		entry.Message,
		metaStr,
	)

	return nil
}

// Helpers
func (c *ConsoleLogger) Info(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return c.Log(ctx, logger.LogEntry{Timestamp: ts, Level: logger.LevelInfo, Message: msg, Metadata: meta})
}

func (c *ConsoleLogger) Error(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return c.Log(ctx, logger.LogEntry{Timestamp: ts, Level: logger.LevelError, Message: msg, Metadata: meta})
}

func (c *ConsoleLogger) Warn(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return c.Log(ctx, logger.LogEntry{Timestamp: ts, Level: logger.LevelWarn, Message: msg, Metadata: meta})
}

func (c *ConsoleLogger) Debug(ctx context.Context, ts time.Time, msg string, meta map[string]interface{}) error {
	return c.Log(ctx, logger.LogEntry{Timestamp: ts, Level: logger.LevelDebug, Message: msg, Metadata: meta})
}

func (c *ConsoleLogger) Close() error {
	return nil
}
