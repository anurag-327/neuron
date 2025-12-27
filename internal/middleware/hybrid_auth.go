package middleware

import (
	"net/http"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

// HybridAuthMiddleware checks for either a valid Bearer Token or a valid X-API-Key credential
func HybridAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")

		var isAuthenticated bool

		// 1. Try Bearer Token
		if authHeader != "" {
			token, ok := util.ValidateHeaderToken(authHeader)
			if ok {
				claims, err := util.ParseJWTToken(token, config.JwtSecret)
				if err == nil {
					user, err := repository.GetUserByID(ctx, claims.Sub)
					if err == nil {
						c.Set("user", user)
						isAuthenticated = true
					}
				}
			}
		}

		// 2. Try API Key if not authenticated yet
		if !isAuthenticated && apiKey != "" {
			cred, err := services.ValidateAPIKey(ctx, apiKey)
			if err == nil {
				user, err := repository.GetUserByID(ctx, cred.UserID)
				if err == nil {
					c.Set("user", user)
					c.Set("credential", cred)
					c.Set("user_id", user.ID)
					isAuthenticated = true
				}
			}
		}

		if isAuthenticated {
			c.Next()
			return
		}

		response.Error(c, http.StatusUnauthorized, "unauthorized: invalid or missing authentication credentials")
		c.Abort()
	}
}
