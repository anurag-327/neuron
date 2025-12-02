package repository

import (
	"context"
	"errors"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo"
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
	}
	return job, nil
}

func DeleteJob(ctx context.Context, job *models.Job) error {
	coll := mgm.Coll(job)
	return coll.DeleteWithCtx(ctx, job)
}
