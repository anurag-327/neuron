package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurag-327/neuron/db"
	"github.com/anurag-327/neuron/internal/factory"
	sandboxUtil "github.com/anurag-327/neuron/internal/util/sandbox"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables")
	}
	db.ConnectMongoDB()
}

func main() {
	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture system signals (Ctrl+C, Docker stop, etc.)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Define topics and handlers
	topics := map[string]struct {
		Group         string
		Handler       func([]byte) error
		MaxConcurrent int
	}{
		"code-jobs": {
			Group:         "code-runner-group",
			Handler:       sandboxUtil.ExecuteCode,
			MaxConcurrent: 1000,
		},
	}

	for topic, cfg := range topics {
		go func(topic, group string, handler func([]byte) error, maxConcurrent int) {
			c := factory.GetSubscriber(group, topic)
			defer c.Close()
			c.ConsumeControlled(ctx, handler, maxConcurrent)
		}(topic, cfg.Group, cfg.Handler, cfg.MaxConcurrent)
	}

	//  Wait for shutdown signal
	<-sigChan
	log.Println("🛑 Shutdown signal received... cleaning up")

	//  Cancel context so all consumers exit gracefully
	cancel()

	// Give consumers time to exit cleanly
	time.Sleep(2 * time.Second)
	log.Println(" All consumers stopped gracefully")
}
