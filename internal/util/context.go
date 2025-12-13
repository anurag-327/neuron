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

func GetPageFromQuery(c *gin.Context) (int, error) {
	pageStr := c.Query("page")
	if pageStr == "" {
		return 1, nil
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return 0, errors.New("page must be a number")
	}
	if page <= 0 {
		return 0, errors.New("page must be greater than 0")
	}
	return page, nil
}

func GetLimitFromQuery(c *gin.Context) (int, error) {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return 10, nil
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, errors.New("limit must be a number")
	}
	if limit <= 0 {
		return 0, errors.New("limit must be greater than 0")
	}
	return limit, nil
}

func GetPageAndLimitFromQuery(c *gin.Context) (int, int, error) {
	page, err := GetPageFromQuery(c)
	if err != nil {
		return 0, 0, err
	}
	limit, err := GetLimitFromQuery(c)
	if err != nil {
		return 0, 0, err
	}
	return page, limit, nil
}
