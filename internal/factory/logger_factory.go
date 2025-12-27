package factory

import (
	"fmt"
	"os"

	"github.com/anurag-327/neuron/pkg/logger"
	consoleLogger "github.com/anurag-327/neuron/pkg/logger/console"
	redisLogger "github.com/anurag-327/neuron/pkg/logger/redis"
)

// GetLogger returns the appropriate logger based on environment
// - Development: Console logger (prints to stdout)
// - Production: Redis logger (pushes to Redis queue)
func GetLogger() (logger.Logger, error) {
	env := os.Getenv("ENV")
	serviceName := os.Getenv("SERVICE_NAME")

	if serviceName == "" {
		serviceName = "neuron-backend"
	}

	// Development environment - use console logger
	if env == "dev" || env == "development" || env == "" {
		return consoleLogger.NewConsoleLogger(serviceName), nil
	}

	// Production environment - use Redis logger
	return createRedisLogger(serviceName)
}

// createRedisLogger creates a Redis-based logger for production
func createRedisLogger(serviceName string) (logger.Logger, error) {
	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDRESS is required for production logging")
	}

	cfg := redisLogger.Config{
		ServiceName: serviceName,
		RedisAddr:   redisAddr,
		RedisUser:   os.Getenv("REDIS_USER"),
		RedisPass:   os.Getenv("REDIS_PASSWORD"),
		RedisDB:     0,
		QueueName:   getLogQueueName(),
	}

	redisLog, err := redisLogger.NewLoggerClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis logger: %w", err)
	}

	return redisLog, nil
}

// getLogQueueName returns the queue name for logs
func getLogQueueName() string {
	queueName := os.Getenv("LOG_QUEUE_NAME")
	if queueName == "" {
		queueName = "neuron_logs_queue"
	}
	return queueName
}

// InitializeGlobalLogger initializes and sets the global logger
// Call this in your main.go init() or main() function
func InitializeGlobalLogger() error {
	log, err := GetLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.SetGlobalLogger(log)
	return nil
}
