package routes

import (
	logsHandler "github.com/anurag-327/neuron/internal/handler/logs"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterLogsRoutes(router *gin.RouterGroup) {
	logsRouter := router.Group("/logs")
	{
		logsRouter.GET("/api", middleware.VerifyTokenMiddleware(), logsHandler.GetApiLogsHandler)
		logsRouter.GET("/jobs", middleware.VerifyTokenMiddleware(), logsHandler.GetJobLogsHandler)
		logsRouter.GET("/credits", middleware.VerifyTokenMiddleware(), logsHandler.GetCreditLogsHandler)
		logsRouter.GET("/recent-activity", middleware.VerifyTokenMiddleware(), logsHandler.GetRecentActivityHandler)
	}
}
