package util

import (
	"errors"
	"strconv"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/gin-gonic/gin"
)

func GetUserFromContext(c *gin.Context) (*models.User, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}
	user, ok := userValue.(*models.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

func GetPageFromQuery(c *gin.Context) int64 {
	pageStr := c.Query("page")
	if pageStr == "" {
		return int64(1)
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return int64(1) // fallback to page 1
	}
	return int64(page)
}

func GetLimitFromQuery(c *gin.Context) int64 {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return int64(10)
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return int64(10) // fallback to default limit
	}
	// enforce maximum limit of 25
	if limit > 100 {
		return int64(100)
	}
	return int64(limit)
}

func GetPageAndLimitFromQuery(c *gin.Context) (int64, int64) {
	page := GetPageFromQuery(c)
	limit := GetLimitFromQuery(c)
	return page, limit
}
