package authHandler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userInfo, err := util.GetGoogleUserInfo(body.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid token or failed to fetch user info")
		return
	}

	email := userInfo.Email

	user, err := repository.GetUserByEmail(ctx, email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get ")
		return
	}

	if user.Role != "admin" {
		response.Error(c, http.StatusForbidden, "You are not authorized to access this resource")
		return
	}

	token, err := util.GenerateJWTToken(user, config.TokenExpirationTime, config.JwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	c.Header("Authorization", token)
	response.Success(c, http.StatusOK, "Google login successful", gin.H{
		"newUser": false,
	})
}

func GoogleLoginInController(c *gin.Context) {
	ctx := c.Request.Context()

	var body struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userInfo, err := util.GetGoogleUserInfo(body.Token, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid token or failed to fetch user info")
		return
	}

	email := userInfo.Email
	profilePhoto := userInfo.Picture
	name := userInfo.Name
	verified := userInfo.Verified

	user, err := repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Email not found, create a new user
			// Generate a random password for the new user
			randomPassword := util.GenerateRandomString(10)
			hashedPassword, err := util.Encrypt(randomPassword)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to encrypt password")
				return
			}

			newUser := &models.User{
				Email:        email,
				Password:     hashedPassword,
				Role:         models.RoleTypeUser,
				Verified:     verified,
				AuthProvider: string(models.AuthProviderGoogle),
				Username:     strings.Split(email, "@")[0],
				Name:         name,
				ImageUrl:     &profilePhoto,
				Credits:      0,
			}

			newUser, err = repository.SaveUser(ctx, newUser)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to create user")
				return
			}

			err = services.CreditUserAndLog(ctx, newUser.ID, models.DefaultSignupCredits, models.CreditReasonSignupBonus, nil, nil)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to credit user")
				return
			}

			token, err := util.GenerateJWTToken(newUser, config.TokenExpirationTime, config.JwtSecret)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to generate token")
				return
			}

			c.Header("Authorization", token)
			response.Success(c, http.StatusOK, "Google login successful", gin.H{
				"newUser": true,
			})
			return
		}

		response.Error(c, http.StatusInternalServerError, "Failed to get user by email")
		return
	}

	token, err := util.GenerateJWTToken(user, config.TokenExpirationTime, config.JwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	c.Header("Authorization", token)
	response.Success(c, http.StatusOK, "Google login successful", gin.H{
		"newUser": false,
	})
}

func GithubLoginInController(c *gin.Context) {
	ctx := c.Request.Context()

	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	token, err := util.GetGitHubAccessToken(body.Code)
	if err != nil {
		fmt.Println("Error getting GitHub access token:", err)
		response.Error(c, http.StatusInternalServerError, "Failed to get access token from GitHub")
		return
	}
	userInfo, err := util.FetchGitHubUserInfo(token)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get user info from GitHub")
		return
	}
	email := userInfo.Email
	profilePhoto := userInfo.AvatarURL
	name := userInfo.Name
	verified := userInfo.Verified

	user, err := repository.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Email not found, create a new user
			// Generate a random password for the new user
			randomPassword := util.GenerateRandomString(10)
			hashedPassword, err := util.Encrypt(randomPassword)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to encrypt password")
				return
			}

			newUser := &models.User{
				Email:        email,
				Password:     hashedPassword,
				Role:         models.RoleTypeUser,
				Verified:     verified,
				AuthProvider: string(models.AuthProviderGoogle),
				Username:     strings.Split(email, "@")[0],
				Name:         name,
				ImageUrl:     &profilePhoto,
				Credits:      0,
			}

			newUser, err = repository.SaveUser(ctx, newUser)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to create user")
				return
			}

			err = services.CreditUserAndLog(ctx, newUser.ID, models.DefaultSignupCredits, models.CreditReasonSignupBonus, nil, nil)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to credit user")
				return
			}

			token, err := util.GenerateJWTToken(newUser, config.TokenExpirationTime, config.JwtSecret)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to generate token")
				return
			}

			c.Header("Authorization", token)
			response.Success(c, http.StatusOK, "Google login successful", gin.H{
				"newUser": true,
			})
			return
		}

		response.Error(c, http.StatusInternalServerError, "Failed to get user by email")
		return
	}

	token, err = util.GenerateJWTToken(user, config.TokenExpirationTime, config.JwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	c.Header("Authorization", token)
	response.Success(c, http.StatusOK, "Github login successful", gin.H{
		"newUser": false,
	})

}

func InitPostmanController(c *gin.Context) {
	ctx := c.Request.Context()

	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	xAdminKey := c.GetHeader("x-admin-key")
	if xAdminKey == "" {
		response.Error(c, http.StatusBadRequest, "Missing Authorization Header")
		c.Abort()
		return
	}

	if xAdminKey != os.Getenv("X_ADMIN_KEY") {
		response.Error(c, http.StatusUnauthorized, "Invalid admin key")
		return
	}

	user, err := repository.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if user.Role != "admin" {
		response.Error(c, http.StatusForbidden, "You are not authorized to access this resource")
		return
	}

	token, err := util.GenerateJWTToken(user, config.TokenExpirationTime, config.JwtSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Authorization", token)
	response.Success(c, http.StatusOK, "Login successful", nil)
}
