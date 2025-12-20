package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrJobNotFound = errors.New("job not found")

func SaveJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	coll := mgm.Coll(job)
	if job.ID.IsZero() {
		// New job → create
		if err := coll.CreateWithCtx(ctx, job); err != nil {
			return nil, err
		}
	} else {
		// Existing job → update
		if err := coll.UpdateWithCtx(ctx, job); err != nil {
			return nil, err
		}
	}
	return job, nil
}

func GetJobByID(ctx context.Context, id string) (*models.Job, error) {
	job := &models.Job{}
	coll := mgm.Coll(job)
	if err := coll.FindByIDWithCtx(ctx, id, job); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	return job, nil
}

func GetJobByIDAndUserID(
	ctx context.Context,
	jobID primitive.ObjectID,
	userID primitive.ObjectID,
) (*models.Job, error) {

	job := &models.Job{}
	coll := mgm.Coll(job)

	err := coll.FindOne(
		ctx,
		bson.M{
			"_id":    jobID,
			"userId": userID,
		},
	).Decode(job)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return job, nil
}

func DeleteJob(ctx context.Context, job *models.Job) error {
	coll := mgm.Coll(job)
	return coll.DeleteWithCtx(ctx, job)
}

func GetJobsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int64) ([]models.Job, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	const maxLimit int64 = 100
	if limit > maxLimit {
		limit = maxLimit
	}

	skip := (page - 1) * limit
	coll := mgm.Coll(&models.Job{})
	cursor, err := coll.Find(
		ctx,
		bson.M{"userId": userID},
		options.Find().
			SetSort(bson.M{"created_at": -1}).
			SetSkip(skip).
			SetLimit(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer cursor.Close(ctx)
	var jobs []models.Job
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}
	return jobs, nil
}

func CountJobsByUserID(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	coll := mgm.Coll(&models.Job{})
	count, err := coll.CountDocuments(ctx, bson.M{"userId": userID})
	if err != nil {
		return 0, fmt.Errorf("failed to count jobs: %w", err)
	}
	return count, nil
}

// GetJobStatsByUserID aggregates job statistics for a user within a date range
func GetJobStatsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	startDate, endDate primitive.DateTime,
) (*models.JobStats, error) {

	coll := mgm.Coll(&models.Job{})
	matchFilter := bson.M{"userId": userID}

	// Date filter
	if startDate > 0 || endDate > 0 {
		dateFilter := bson.M{}
		if startDate > 0 {
			dateFilter["$gte"] = startDate
		}
		if endDate > 0 {
			dateFilter["$lte"] = endDate
		}
		// Assuming we care about when the job finished for stats
		matchFilter["finishedAt"] = dateFilter
	}
	// Also ensure job is finished
	matchFilter["status"] = bson.M{"$in": bson.A{models.StatusSuccess, models.StatusFailed, models.StatusQueued, models.StatusRunning}}
	// Actually for "executed" stats we probably only want Success/Failed
	matchFilter["status"] = bson.M{"$in": bson.A{models.StatusSuccess, models.StatusFailed}}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":               nil,
			"totalExecutedJobs": bson.M{"$sum": 1},
			"successJobs": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$status", models.StatusSuccess}},
						1,
						0,
					},
				},
			},
			"failedJobs": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$status", models.StatusFailed}},
						1,
						0,
					},
				},
			},
			"avgExecutionMs": bson.M{
				"$avg": bson.M{"$subtract": bson.A{"$finishedAt", "$startedAt"}},
			},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate job stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalExecutedJobs int64   `bson:"totalExecutedJobs"`
		SuccessJobs       int64   `bson:"successJobs"`
		FailedJobs        int64   `bson:"failedJobs"`
		AvgExecutionMs    float64 `bson:"avgExecutionMs"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode job stats: %w", err)
	}

	stats := &models.JobStats{}
	if len(results) > 0 {
		stats.TotalExecutedJobs = results[0].TotalExecutedJobs
		stats.SuccessJobs = results[0].SuccessJobs
		stats.FailedJobs = results[0].FailedJobs
		stats.AvgExecutionMs = results[0].AvgExecutionMs
	}

	return stats, nil
}
