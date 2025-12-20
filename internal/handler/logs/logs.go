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

	// Fetch API logs
	apiLogs, err := repository.GetApiLogsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch total count
	total, err := repository.CountApiLogsByUserID(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "api logs fetched successfully", gin.H{
		"apiLogs": apiLogs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func GetJobLogsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := util.GetPageAndLimitFromQuery(c)

	// Fetch job logs
	jobLogs, err := repository.GetJobsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch total count
	total, err := repository.CountJobsByUserID(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "job logs fetched successfully", gin.H{
		"jobLogs": jobLogs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

func GetCreditLogsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := util.GetPageAndLimitFromQuery(c)

	// Fetch credit logs
	creditLogs, err := repository.GetCreditTransactionsByUserID(ctx, user.ID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch total count
	total, err := repository.CountCreditTransactionsByUserID(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "credit logs fetched successfully", gin.H{
		"creditLogs": creditLogs,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

func GetRecentActivityHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get recent 7 API logs without date filtering
	apiLogs, err := repository.GetApiLogsByUserID(ctx, user.ID, 1, 7)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "recent activity fetched successfully", gin.H{"apiLogs": apiLogs})
}
