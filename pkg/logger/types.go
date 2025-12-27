package logger

import (
	"time"
)

// LogLevel defines the severity of the log
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)


// LogEntry is the immutable log record.
// This struct captures all necessary context for a log event.
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
