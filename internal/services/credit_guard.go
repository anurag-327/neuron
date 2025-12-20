package services

import (
	"context"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AssertCanSubmit(
	ctx context.Context,
	userID primitive.ObjectID,
) error {
	amount := config.GetCreditsForReason(models.CreditReasonSubmission)
	return repository.HasSufficientCredits(ctx, userID, amount)
}
