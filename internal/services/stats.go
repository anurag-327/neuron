package services

import (
	"context"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStats struct {
	ApiStats *models.ApiStats `json:"apiStats"`
	JobStats *models.JobStats `json:"jobStats"`
	Credits  int64            `json:"credits"`
}

// GetUserStats retrieves comprehensive statistics for a user (API, Jobs, Credits)
// Currently aggregates all-time statistics (no date filtering)
func GetUserStats(ctx context.Context, userID primitive.ObjectID) (*UserStats, error) {
	// Pass zero dates to fetch all-time stats
	var zeroDate primitive.DateTime

	apiStats, err := repository.GetApiStatsByUserID(ctx, userID, zeroDate, zeroDate)
	if err != nil {
		return nil, err
	}

	jobStats, err := repository.GetJobStatsByUserID(ctx, userID, zeroDate, zeroDate)
	if err != nil {
		return nil, err
	}

	user, err := repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserStats{
		ApiStats: apiStats,
		JobStats: jobStats,
		Credits:  user.Credits,
	}, nil
}
