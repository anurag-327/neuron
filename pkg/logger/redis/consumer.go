package redisLogger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// redisConsumer implements Consumer interface
type redisConsumer struct {
	client *redis.Client
	queue  string
}

// NewConsumer creates a new Redis-backed Consumer.
func NewConsumer(cfg Config) (logger.Consumer, error) {
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("redis address is required (e.g. localhost:6379)")
	}

	if cfg.QueueName == "" {
		cfg.QueueName = "logs_queue"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUser,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	//  Ping to verify
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("consumer failed to connect to redis: %w", err)
	}

	return &redisConsumer{
		client: rdb,
		queue:  cfg.QueueName,
	}, nil
}

// FetchBatch retrieves logs using a "Greedy then Block" strategy to minimize latency.
func (c *redisConsumer) FetchBatch(ctx context.Context, batchSize int, timeout time.Duration) ([]logger.LogEntry, error) {
	if batchSize <= 0 {
		batchSize = 1000
	}

	// 1. Attempt to fetch 'batchSize' items immediately (Non-blocking)
	// LPopCount is efficient for grabbing multiple items.
	results, err := c.client.LPopCount(ctx, c.queue, batchSize).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis lpop error: %w", err)
	}

	// 2. If queue was empty, switch to Blocking mode (Long Polling)
	if len(results) == 0 {
		// BLPop waits until at least one item arrives or timeout occurs.
		// Returns [queue_name, value]
		res, err := c.client.BLPop(ctx, timeout, c.queue).Result()
		if err != nil {
			if err == redis.Nil {
				// Timeout reached, simply return empty batch (no error)
				return []logger.LogEntry{}, nil
			}
			return nil, fmt.Errorf("redis blpop error: %w", err)
		}

		// We got one item!
		results = append(results, res[1])

		// Optimization: Since we woke up, there might be more items queued behind it (burst traffic).
		// Try to fill the rest of the batch immediately.
		remaining := batchSize - 1
		if remaining > 0 {
			more, err := c.client.LPopCount(ctx, c.queue, remaining).Result()
			if err == nil {
				results = append(results, more...)
			}
		}
	}

	// 3. Deserialize JSON strings to LogEntry structs
	logs := make([]logger.LogEntry, 0, len(results))
	for _, data := range results {
		var l logger.LogEntry
		// We ignore malformed logs to prevent poison pill crashing the worker
		// optionally: log this error to stderr
		if err := json.Unmarshal([]byte(data), &l); err == nil {
			logs = append(logs, l)
		}
	}

	return logs, nil
}

func (c *redisConsumer) Close() error {
	return c.client.Close()
}
