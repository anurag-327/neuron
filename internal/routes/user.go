package routes

import (
	userHandler "github.com/anurag-327/neuron/internal/handler/user"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup) {
	authRouter := router.Group("/user")
	{
		authRouter.GET("/me", middleware.VerifyTokenMiddleware(), userHandler.GetUserHandler)
		authRouter.GET("/stats", middleware.VerifyTokenMiddleware(), userHandler.GetUserStatsHandler)
	}
}
