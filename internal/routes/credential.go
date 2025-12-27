package routes

import (
	credentialHandler "github.com/anurag-327/neuron/internal/handler/credential"
	"github.com/anurag-327/neuron/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterCredentialRoutes(router *gin.RouterGroup) {
	credRouter := router.Group("/credentials")
	credRouter.Use(middleware.VerifyTokenMiddleware()) // Manage credentials using Token Auth
	{
		credRouter.POST("/create", credentialHandler.CreateCredentialHandler)
		credRouter.GET("/get", credentialHandler.GetCredentialHandler)
		credRouter.POST("/reveal", credentialHandler.RevealCredentialHandler)
		credRouter.DELETE("/revoke", credentialHandler.RevokeCredentialHandler)
	}
}
