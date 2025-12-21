package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetTotalExecutions returns the count of all jobs for a user in the last 30 days
func GetTotalExecutions(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	coll := mgm.Coll(&models.Job{})
	last30Days := time.Now().AddDate(0, 0, -30)

	count, err := coll.CountDocuments(ctx, bson.M{
		"userId":     userID,
		"created_at": bson.M{"$gte": last30Days},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count total executions: %w", err)
	}
	return count, nil
}

// GetSuccessRate calculates the success rate for a user
func GetSuccessRate(ctx context.Context, userID primitive.ObjectID) (float64, error) {
	coll := mgm.Coll(&models.Job{})
	last30Days := time.Now().AddDate(0, 0, -30)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"created_at": bson.M{"$gte": last30Days},
			"status":     bson.M{"$in": bson.A{models.StatusSuccess, models.StatusFailed}},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": 1},
			"successful": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{
							"$and": bson.A{
								bson.M{"$eq": bson.A{"$status", models.StatusSuccess}},
								bson.M{"$not": bson.A{
									bson.M{"$in": bson.A{"$errorType", bson.A{models.ErrSandboxError, models.ErrInternalError}}},
								}},
							},
						},
						1,
						0,
					},
				},
			},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate success rate: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total      int64 `bson:"total"`
		Successful int64 `bson:"successful"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, fmt.Errorf("failed to decode success rate: %w", err)
	}

	if len(results) == 0 || results[0].Total == 0 {
		return 0, nil
	}

	return (float64(results[0].Successful) / float64(results[0].Total)) * 100, nil
}

// GetAvgResponseTime calculates average response time in milliseconds, excluding outliers
func GetAvgResponseTime(ctx context.Context, userID primitive.ObjectID) (float64, error) {
	coll := mgm.Coll(&models.Job{})
	last30Days := time.Now().AddDate(0, 0, -30)

	// First, get all response times
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"created_at": bson.M{"$gte": last30Days},
			"status":     models.StatusSuccess,
			"errorType":  bson.M{"$nin": bson.A{models.ErrSandboxError, models.ErrInternalError}},
			"startedAt":  bson.M{"$exists": true},
			"finishedAt": bson.M{"$exists": true},
		}}},
		{{Key: "$project", Value: bson.M{
			"responseTime": bson.M{"$subtract": bson.A{"$finishedAt", "$startedAt"}},
		}}},
		{{Key: "$sort", Value: bson.M{"responseTime": 1}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("failed to get response times: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		ResponseTime float64 `bson:"responseTime"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return 0, fmt.Errorf("failed to decode response times: %w", err)
	}

	if len(results) == 0 {
		return 0, nil
	}

	// Extract response times into a slice
	times := make([]float64, len(results))
	for i, r := range results {
		times[i] = r.ResponseTime
	}

	// Filter outliers using IQR method
	filteredTimes := filterOutliers(times)

	if len(filteredTimes) == 0 {
		return 0, nil
	}

	// Calculate average of filtered times
	var sum float64
	for _, t := range filteredTimes {
		sum += t
	}

	return sum / float64(len(filteredTimes)), nil
}

// filterOutliers removes outliers using the Interquartile Range (IQR) method
// Values below Q1 - 1.5*IQR or above Q3 + 1.5*IQR are considered outliers
func filterOutliers(data []float64) []float64 {
	if len(data) < 4 {
		// Not enough data points for IQR, return as is
		return data
	}

	// Data is already sorted from the MongoDB query
	n := len(data)

	// Calculate Q1 (25th percentile) and Q3 (75th percentile)
	q1Index := n / 4
	q3Index := (3 * n) / 4

	q1 := data[q1Index]
	q3 := data[q3Index]

	// Calculate IQR
	iqr := q3 - q1

	// Calculate bounds
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr

	// Filter data
	filtered := make([]float64, 0, n)
	for _, value := range data {
		if value >= lowerBound && value <= upperBound {
			filtered = append(filtered, value)
		}
	}

	return filtered
}

// LanguageUsageStats represents language usage statistics
type LanguageUsageStats struct {
	Language   string `bson:"_id" json:"language"`
	Count      int64  `bson:"count" json:"count"`
	Percentage int    `json:"percentage"`
}

// GetLanguageUsage returns top 4 languages by usage
func GetLanguageUsage(ctx context.Context, userID primitive.ObjectID) ([]LanguageUsageStats, error) {
	coll := mgm.Coll(&models.Job{})
	last30Days := time.Now().AddDate(0, 0, -30)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"created_at": bson.M{"$gte": last30Days},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$language",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 4}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get language usage: %w", err)
	}
	defer cursor.Close(ctx)

	var results []LanguageUsageStats
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode language usage: %w", err)
	}

	// Calculate total for percentages
	var total int64
	for _, r := range results {
		total += r.Count
	}

	// Calculate percentages
	for i := range results {
		if total > 0 {
			results[i].Percentage = int((float64(results[i].Count) / float64(total)) * 100)
		}
	}

	return results, nil
}

// WeeklyTrendStats represents daily execution statistics
type WeeklyTrendStats struct {
	Day        string `json:"day"`
	Date       string `json:"date"`
	Executions int64  `json:"executions"`
}

// GetWeeklyTrend returns execution counts for the last 7 days
func GetWeeklyTrend(ctx context.Context, userID primitive.ObjectID) ([]WeeklyTrendStats, error) {
	coll := mgm.Coll(&models.Job{})

	// Get last 7 days
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -6) // -6 because we include today

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"userId":     userID,
			"created_at": bson.M{"$gte": sevenDaysAgo},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$created_at",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly trend: %w", err)
	}
	defer cursor.Close(ctx)

	var dbResults []struct {
		Date  string `bson:"_id"`
		Count int64  `bson:"count"`
	}

	if err := cursor.All(ctx, &dbResults); err != nil {
		return nil, fmt.Errorf("failed to decode weekly trend: %w", err)
	}

	// Create a map for quick lookup
	countMap := make(map[string]int64)
	for _, r := range dbResults {
		countMap[r.Date] = r.Count
	}

	// Generate all 7 days
	results := make([]WeeklyTrendStats, 7)
	dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	for i := 0; i < 7; i++ {
		date := sevenDaysAgo.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")

		results[i] = WeeklyTrendStats{
			Day:        dayNames[date.Weekday()],
			Date:       dateStr,
			Executions: countMap[dateStr],
		}
	}

	return results, nil
}

// InsightsStats represents performance insights
type InsightsStats struct {
	TopLanguage           string `json:"topLanguage"`
	TopLanguagePercentage int    `json:"topLanguagePercentage"`
	PeakUsageDay          string `json:"peakUsageDay"`
	PeakUsageTime         string `json:"peakUsageTime"`
}

// GetInsights calculates insights from the data
func GetInsights(ctx context.Context, userID primitive.ObjectID, languageUsage []LanguageUsageStats, weeklyTrend []WeeklyTrendStats) *InsightsStats {
	insights := &InsightsStats{
		PeakUsageTime: "afternoons", // Static for now
	}

	// Top language
	if len(languageUsage) > 0 {
		insights.TopLanguage = languageUsage[0].Language
		insights.TopLanguagePercentage = languageUsage[0].Percentage
	}

	// Peak usage day
	var maxExecutions int64
	var peakDay string
	for _, day := range weeklyTrend {
		if day.Executions > maxExecutions {
			maxExecutions = day.Executions
			peakDay = day.Day
		}
	}
	insights.PeakUsageDay = peakDay

	return insights
}
