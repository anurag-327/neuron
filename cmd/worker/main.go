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
	"github.com/anurag-327/neuron/pkg/messaging"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found, using environment variables")
	}
	db.ConnectMongoDB()
}

type ConfigStruct struct {
	Group         string
	Handler       func([]byte) error
	MaxConcurrent int
}

func main() {
	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture system signals (Ctrl+C, Docker stop, etc.)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Define topics and handlers
	topics := map[string]ConfigStruct{
		"code-jobs": {
			Group:         "code-runner-group",
			Handler:       sandboxUtil.ExecuteCode,
			MaxConcurrent: 1000,
		},
	}

	for topic, cfg := range topics {
		// initialize subscriber BEFORE starting goroutine
		sub, err := factory.GetSubscriber(cfg.Group, topic)
		if err != nil {
			log.Fatalf("Failed to initialize subscriber for %s: %v", topic, err)
		}

		go func(sub messaging.Subscriber, cfg ConfigStruct) {
			defer sub.Close()
			sub.ConsumeControlled(ctx, cfg.Handler, cfg.MaxConcurrent)
		}(sub, cfg)
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
