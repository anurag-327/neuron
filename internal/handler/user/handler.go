package userHandler

import (
	"net/http"

	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context){
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to get user from context")
		return
	}
	response.Success(c,http.StatusOK,"user fetched successfully",gin.H{
		"user":user,
	})
}