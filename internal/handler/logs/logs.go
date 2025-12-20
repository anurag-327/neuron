package logsHandler

import (
	"net/http"

	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func GetApiLogsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := util.GetPageAndLimitFromQuery(c)

	apiLogs, err := repository.GetApiLogsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "api logs fetched successfully", gin.H{"apiLogs": apiLogs})
}

func GetJobLogsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := util.GetPageAndLimitFromQuery(c)

	jobLogs, err := repository.GetJobsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "job logs fetched successfully", gin.H{"jobLogs": jobLogs})
}

func GetCreditLogsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := util.GetPageAndLimitFromQuery(c)

	creditLogs, err := repository.GetCreditTransactionsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "credit logs fetched successfully", gin.H{"creditLogs": creditLogs})
}
