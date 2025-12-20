package main

import (
	"log"
	"net/http"
	"os"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/anurag-327/neuron/internal/models"
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
	config.JwtSecret = []byte(os.Getenv("JWT_SECRET"))
	conn.ConnectMongoDB()
	models.CreateUserIndexes()
	models.CreateCreditTransactionIndexes()
	models.CreateJobIndexes()
	models.CreateApiLogIndexes()
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

	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, "healthy")
	})
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		response.Error(c, http.StatusNotFound, "Route '"+path+"' does not exist. Please check the API documentation.")
	})

	router.Run(":" + os.Getenv("PORT"))
}
