package repository

import (
	"context"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ApiLogStats represents statistics for API logs
type ApiLogStats struct {
	Total                int       `json:"total"`
	AverageQueueTime     float64   `json:"averageQueueTime"`
	AverageExecutionTime float64   `json:"averageExecutionTime"`
	QueueTimes           []float64 `json:"queueTimes"`
	ExecutionTimes       []float64 `json:"executionTimes"`
}

// GetApiLogStats calculates statistics for API logs
func GetApiLogStats(ctx context.Context, userID primitive.ObjectID) (*ApiLogStats, error) {
	coll := mgm.Coll(&models.ApiLog{})

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"status":     models.StatusSuccess,
			"startedAt":  bson.M{"$exists": true},
			"finishedAt": bson.M{"$exists": true},
			"queuedAt":   bson.M{"$exists": true},
		}}},
		{{Key: "$project", Value: bson.M{
			"queueTime":     bson.M{"$subtract": bson.A{"$startedAt", "$queuedAt"}},
			"executionTime": bson.M{"$subtract": bson.A{"$finishedAt", "$startedAt"}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get api log stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		QueueTime     float64 `bson:"queueTime"`
		ExecutionTime float64 `bson:"executionTime"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode api log stats: %w", err)
	}

	stats := &ApiLogStats{
		Total:          len(results),
		QueueTimes:     make([]float64, len(results)),
		ExecutionTimes: make([]float64, len(results)),
	}

	if len(results) == 0 {
		return stats, nil
	}

	// Extract times
	for i, r := range results {
		stats.QueueTimes[i] = r.QueueTime
		stats.ExecutionTimes[i] = r.ExecutionTime
	}

	// Calculate averages excluding outliers
	stats.AverageQueueTime = calculateAverageWithoutOutliers(stats.QueueTimes)
	stats.AverageExecutionTime = calculateAverageWithoutOutliers(stats.ExecutionTimes)

	return stats, nil
}

// JobLogStats represents statistics for job logs
type JobLogStats struct {
	Total                int       `json:"total"`
	AverageQueueTime     float64   `json:"averageQueueTime"`
	AverageExecutionTime float64   `json:"averageExecutionTime"`
	QueueTimes           []float64 `json:"queueTimes"`
	ExecutionTimes       []float64 `json:"executionTimes"`
}

// GetJobLogStats calculates statistics for job logs
func GetJobLogStats(ctx context.Context, userID primitive.ObjectID) (*JobLogStats, error) {
	coll := mgm.Coll(&models.Job{})

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"status":     models.StatusSuccess,
			"startedAt":  bson.M{"$exists": true},
			"finishedAt": bson.M{"$exists": true},
			"queuedAt":   bson.M{"$exists": true},
		}}},
		{{Key: "$project", Value: bson.M{
			"queueTime":     bson.M{"$subtract": bson.A{"$startedAt", "$queuedAt"}},
			"executionTime": bson.M{"$subtract": bson.A{"$finishedAt", "$startedAt"}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get job log stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		QueueTime     float64 `bson:"queueTime"`
		ExecutionTime float64 `bson:"executionTime"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode job log stats: %w", err)
	}

	stats := &JobLogStats{
		Total:          len(results),
		QueueTimes:     make([]float64, len(results)),
		ExecutionTimes: make([]float64, len(results)),
	}

	if len(results) == 0 {
		return stats, nil
	}

	// Extract times
	for i, r := range results {
		stats.QueueTimes[i] = r.QueueTime
		stats.ExecutionTimes[i] = r.ExecutionTime
	}

	// Calculate averages excluding outliers
	stats.AverageQueueTime = calculateAverageWithoutOutliers(stats.QueueTimes)
	stats.AverageExecutionTime = calculateAverageWithoutOutliers(stats.ExecutionTimes)

	return stats, nil
}

// calculateAverageWithoutOutliers calculates average after filtering outliers using IQR
func calculateAverageWithoutOutliers(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	// Sort the data
	sorted := make([]float64, len(data))
	copy(sorted, data)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Filter outliers
	filtered := filterOutliersFromSorted(sorted)

	if len(filtered) == 0 {
		return 0
	}

	// Calculate average
	var sum float64
	for _, v := range filtered {
		sum += v
	}

	return sum / float64(len(filtered))
}

// filterOutliersFromSorted removes outliers from already sorted data
func filterOutliersFromSorted(data []float64) []float64 {
	if len(data) < 4 {
		return data
	}

	n := len(data)
	q1Index := n / 4
	q3Index := (3 * n) / 4

	q1 := data[q1Index]
	q3 := data[q3Index]
	iqr := q3 - q1

	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	filtered := make([]float64, 0, n)
	for _, value := range data {
		if value >= lowerBound && value <= upperBound {
			filtered = append(filtered, value)
		}
	}

	return filtered
}
