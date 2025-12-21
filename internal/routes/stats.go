package routes

import (
	"github.com/anurag-327/neuron/internal/handler/stats"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterStatsRoutes(router *gin.RouterGroup) {
	router.GET("/stats", middleware.HybridAuthMiddleware(), stats.GetStats)
}
