package repository

import (
	"context"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveApiLog(ctx context.Context, log *models.ApiLog) (*models.ApiLog, error) {
	coll := mgm.Coll(log)

	if log.ID.IsZero() {
		if err := coll.CreateWithCtx(ctx, log); err != nil {
			return nil, err
		}
	} else {
		if err := coll.UpdateWithCtx(ctx, log); err != nil {
			return nil, err
		}
	}

	return log, nil
}

// GetApiStatsByUserID aggregates API request statistics for a user within a date range
func GetApiStatsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	startDate, endDate primitive.DateTime,
) (*models.ApiStats, error) {

	coll := mgm.Coll(&models.ApiLog{})

	// Build match filter
	matchFilter := bson.M{"userId": userID}
	if startDate > 0 || endDate > 0 {
		dateFilter := bson.M{}
		if startDate > 0 {
			dateFilter["$gte"] = startDate
		}
		if endDate > 0 {
			dateFilter["$lte"] = endDate
		}
		matchFilter["created_at"] = dateFilter
	}

	// Aggregation pipeline for main stats
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":           nil,
			"totalRequests": bson.M{"$sum": 1},
			"successRequests": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$requestStatus", "success"}},
						1,
						0,
					},
				},
			},
			"failedRequests": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$requestStatus", "rejected"}},
						1,
						0,
					},
				},
			},
			"errorRequests": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$requestStatus", "error"}},
						1,
						0,
					},
				},
			},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate api stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalRequests   int64 `bson:"totalRequests"`
		SuccessRequests int64 `bson:"successRequests"`
		FailedRequests  int64 `bson:"failedRequests"`
		ErrorRequests   int64 `bson:"errorRequests"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode api stats: %w", err)
	}

	stats := &models.ApiStats{
		ByEndpoint: make(map[string]int64),
		ByStatus:   make(map[string]int64),
	}

	if len(results) > 0 {
		stats.TotalRequests = results[0].TotalRequests
		stats.SuccessRequests = results[0].SuccessRequests
		stats.FailedRequests = results[0].FailedRequests
		stats.ErrorRequests = results[0].ErrorRequests

		// Calculate success rate
		if stats.TotalRequests > 0 {
			stats.SuccessRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
		}
	}

	return stats, nil
}

func GetApiLogsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	page, limit int64,
) ([]models.ApiLog, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	skip := (page - 1) * limit
	coll := mgm.Coll(&models.ApiLog{})

	cursor, err := coll.Find(
		ctx,
		bson.M{"userId": userID},
		options.Find().
			SetSort(bson.M{"created_at": -1}).
			SetLimit(limit).
			SetSkip(skip),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch api logs: %w", err)
	}

	defer cursor.Close(ctx)

	var logs []models.ApiLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode api logs: %w", err)
	}

	return logs, nil
}

func CountApiLogsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
) (int64, error) {
	coll := mgm.Coll(&models.ApiLog{})
	count, err := coll.CountDocuments(ctx, bson.M{"userId": userID})
	if err != nil {
		return 0, fmt.Errorf("failed to count api logs: %w", err)
	}
	return count, nil
}
