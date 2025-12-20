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
		options.Find().SetSkip(skip).SetLimit(limit),
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
