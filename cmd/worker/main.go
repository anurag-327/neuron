package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/pkg/logger"
	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/anurag-327/neuron/pkg/sandbox/docker"
	"github.com/anurag-327/neuron/pkg/sandbox/docker/pool"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize logger
	if err := factory.InitializeGlobalLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	conn.ConnectMongoDB()
}

func main() {
	appLogger := logger.GetGlobalLogger()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Warm up pools
	if err := docker.InitDockerPool(ctx); err != nil {
		appLogger.Error(ctx, time.Now(), "Pool warm-up failed", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Pool warm-up failed: %v", err)
	}

	// Start consumer worker
	if err := factory.StartConsumer(ctx, config.ExecutionTasksTopic, config.CodeRunnerConsumerGroup, 1000, sandbox.ExecuteCode); err != nil {
		appLogger.Error(ctx, time.Now(), "Failed to start consumer", map[string]interface{}{
			"topic":          config.ExecutionTasksTopic,
			"consumer_group": config.CodeRunnerConsumerGroup,
			"error":          err.Error(),
		})
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received... cleaning up")

	// Destroy all warm containers before exit
	pool.Manager.DestroyAll()

	cancel()

	// Allow clean exit
	time.Sleep(10 * time.Second)
	log.Println("All consumers stopped gracefully")

}
