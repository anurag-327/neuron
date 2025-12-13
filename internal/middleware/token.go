package middleware

import (
	"net/http"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func VerifyTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Missing Authorization Header"))
			c.Abort()
			return
		}

		token, ok := util.ValidateHeaderToken(authHeader)
		if !ok {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : Failed to validate token"))
			c.Abort()
			return
		}

		claims, err := util.ParseJWTToken(token, config.JwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : Failed to parse token"))
			c.Abort()
			return
		}

		user, err := repository.GetUserByID(ctx, claims.Sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : User not found"))
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func OptionalVerifyTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.Next()
			return
		}

		token, ok := util.ValidateHeaderToken(authHeader)
		if !ok {
			c.Next()
			return
		}

		claims, err := util.ParseJWTToken(token, config.JwtSecret)
		if err != nil {
			c.Next()
			return
		}

		user, err := repository.GetUserByID(ctx, claims.Sub)
		if err == nil {
			c.Set("user", user)
		}

		c.Next()
	}
}

func VerifyAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Missing Authorization Header"))
			c.Abort()
			return
		}

		token, ok := util.ValidateHeaderToken(authHeader)
		if !ok {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : Failed to validate token"))
			c.Abort()
			return
		}

		claims, err := util.ParseJWTToken(token, config.JwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : Failed to parse token"))
			c.Abort()
			return
		}

		user, err := repository.GetUserByID(ctx, claims.Sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : User not found"))
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponseUtil(http.StatusUnauthorized, "Unauthorized : Not an admin"))
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
