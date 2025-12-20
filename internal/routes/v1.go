package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterV1Group(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	RegisterAuthRoutes(v1)
	RegisterRunnerRoutes(v1)
	RegisterUserRoutes(v1)
	RegisterLogsRoutes(v1)
}
