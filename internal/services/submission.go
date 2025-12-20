package services

import (
	"context"
	"time"

	"github.com/anurag-327/neuron/internal/dto"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
)

func CreateSubmission(
	ctx context.Context,
	user *models.User,
	body dto.SubmitCodeBody,
) (*models.Job, error) {

	now := time.Now()

	job := &models.Job{
		Language: body.Language,
		Code:     body.Code,
		Input:    body.Input,
		Status:   models.StatusQueued,
		QueuedAt: now,
		UserID:   user.ID,
	}

	return repository.SaveJob(ctx, job)
}
