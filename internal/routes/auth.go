package routes

import (
	authHandler "github.com/anurag-327/neuron/internal/handler/auth"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup) {
	authRouter := router.Group("/auth")
	authRouter.Use(middleware.AuthRateLimit())
	{
		authRouter.POST("/login/admin", authHandler.AdminLogin)
		authRouter.POST("/google", authHandler.GoogleLoginInController)
		authRouter.POST("/github", authHandler.GithubLoginInController)
		authRouter.POST("/init-postman", authHandler.InitPostmanController)
	}
}
