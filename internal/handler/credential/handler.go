package credentialHandler

import (
	"net/http"

	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

// CreateCredentialHandler generates a new API key
func CreateCredentialHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	cred, err := services.CreateCredential(ctx, user.ID)
	if err != nil {
		if err == services.ErrCredentialAlreadyExists {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "credential created successfully", gin.H{
		"credential": cred,
	})
}

// GetCredentialHandler fetches the existing API key
func GetCredentialHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	cred, err := services.GetCredential(ctx, user.ID)
	if err != nil {
		if err == repository.ErrCredentialNotFound {
			response.Error(c, http.StatusNotFound, "no active credential found")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "credential fetched successfully", gin.H{
		"credential": cred,
	})
}

// RevokeCredentialHandler deletes the API key
func RevokeCredentialHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := services.RevokeCredential(ctx, user.ID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "credential revoked successfully", nil)
}
