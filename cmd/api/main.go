package main

import (
	"log"
	"net/http"
	"os"

	"github.com/anurag-327/neuron/db"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/handler"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db.ConnectMongoDB()
}

func main() {
	p := factory.GetPublisher()
	defer p.Close()
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	router.GET("/api/v1/runner/submit", handler.SubmitCodeHandler)
	router.GET("/api/v1/runner/:jobId/status", handler.GetJobStatusHandler)

	router.GET("/health", func(c *gin.Context) {
		response.JSON(c, http.StatusOK, "healthy")
	})
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		response.Error(c, http.StatusNotFound, "Route '"+path+"' does not exist. Please check the API documentation.")
	})

	router.Run(":" + os.Getenv("PORT"))
}
