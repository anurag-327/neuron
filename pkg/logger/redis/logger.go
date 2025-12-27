package redisLogger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type redisLogger struct {
	client      *redis.Client
	queueName   string
	serviceName string
}

func NewLoggerClient(cfg Config) (logger.Logger, error) {
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("redis address is required (e.g. localhost:6379)")
	}
	if cfg.QueueName == "" {
		cfg.QueueName = "logs_queue"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "backend-service"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUser,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
		Protocol: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.RedisAddr, err)
	}

	return &redisLogger{
		client:      client,
		queueName:   cfg.QueueName,
		serviceName: cfg.ServiceName,
	}, nil
}

func (l *redisLogger) Log(ctx context.Context, entry logger.LogEntry) error {
	// Set default values if not provided
	if entry.Service == "" {
		entry.Service = l.serviceName
	}

	if entry.Level == "" {
		entry.Level = logger.LevelInfo
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Serialize log entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Push to Redis (LPush puts it at the head)
	// We use Context from caller to allow timeouts/cancellation
	return l.client.LPush(ctx, l.queueName, data).Err()
}

// Info helper
func (l *redisLogger) Info(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error {
	return l.Log(ctx, logger.LogEntry{Level: logger.LevelInfo, Message: msg, Timestamp: Timestamp, Metadata: meta})
}

// Error helper
func (l *redisLogger) Error(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error {
	return l.Log(ctx, logger.LogEntry{Level: logger.LevelError, Message: msg, Timestamp: Timestamp, Metadata: meta})
}

// Warn helper
func (l *redisLogger) Warn(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error {
	return l.Log(ctx, logger.LogEntry{Level: logger.LevelWarn, Message: msg, Timestamp: Timestamp, Metadata: meta})
}

// Debug helper
func (l *redisLogger) Debug(ctx context.Context, Timestamp time.Time, msg string, meta map[string]interface{}) error {
	return l.Log(ctx, logger.LogEntry{Level: logger.LevelDebug, Message: msg, Timestamp: Timestamp, Metadata: meta})
}

func (l *redisLogger) Close() error {
	return l.client.Close()
}
