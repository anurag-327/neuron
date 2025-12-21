package routes

import (
	runnerHandler "github.com/anurag-327/neuron/internal/handler/runner"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRunnerRoutes(router *gin.RouterGroup) {
	runnerRouter := router.Group("/runner")
	{
		runnerRouter.POST("/submit", middleware.HybridAuthMiddleware(), runnerHandler.SubmitCodeHandler)
		runnerRouter.GET("/:jobId/result", middleware.HybridAuthMiddleware(), runnerHandler.GetJobStatusHandler)
	}
}
