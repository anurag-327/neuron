package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/handler/status"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/anurag-327/neuron/internal/routes"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize logger first
	if err := factory.InitializeGlobalLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Validate JWT_SECRET
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters for security")
	}
	config.JwtSecret = []byte(jwtSecret)

	conn.ConnectMongoDB()
}

func main() {
	publisher, err := factory.GetPublisher()
	if err != nil {
		log.Fatalf("Failed to initialize publisher: %v", err)
	}
	defer publisher.Close()

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	routes.RegisterV1Group(router)

	router.GET("/status", status.GetStatus)
	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, "healthy")
	})
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		response.Error(c, http.StatusNotFound, "Route '"+path+"' does not exist. Please check the API documentation.")
	})

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("API Server starting on port %s", os.Getenv("PORT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received, gracefully shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown completed gracefully")
	}
}
