package userHandler

import (
	"net/http"

	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context) {
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to get user from context")
		return
	}
	response.Success(c, http.StatusOK, "user fetched successfully", gin.H{
		"user": user,
	})
}

func GetUserStatsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := services.GetUserStats(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "user stats fetched successfully", gin.H{
		"stats": stats,
	})
}
