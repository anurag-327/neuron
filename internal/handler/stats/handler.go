package stats

import (
	"net/http"

	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func GetStats(c *gin.Context) {
	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := c.Request.Context()

	totalExecutions, err := repository.GetTotalExecutions(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get total executions")
		return
	}

	successRate, err := repository.GetSuccessRate(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get success rate")
		return
	}

	avgResponseTime, err := repository.GetAvgResponseTime(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get avg response time")
		return
	}

	languageUsage, err := repository.GetLanguageUsage(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get language usage")
		return
	}

	weeklyTrend, err := repository.GetWeeklyTrend(ctx, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get weekly trend")
		return
	}

	insights := repository.GetInsights(ctx, user.ID, languageUsage, weeklyTrend)

	creditsConsumed := int64(0)
	if totalExecutions > 0 {
		creditsConsumed = -totalExecutions
	}

	successRateRounded := float64(int(successRate*10)) / 10

	data := gin.H{
		"summary": gin.H{
			"totalExecutions":  totalExecutions,
			"successRate":      successRateRounded,
			"avgResponseTime":  int(avgResponseTime),
			"creditsRemaining": user.Credits,
			"creditsChange":    creditsConsumed,
		},
		"languageUsage": languageUsage,
		"weeklyTrend":   weeklyTrend,
		"insights":      insights,
	}

	response.Success(c, http.StatusOK, "stats fetched successfully", data)
}
