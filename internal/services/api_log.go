package services

import (
	"context"
	"errors"
	"time"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateApiLog(
	ctx context.Context,
	jobID primitive.ObjectID,
	status models.RunStatus,
	errType *models.SandboxError,
	errMsg string,
	startedAt time.Time,
	finishedAt time.Time,
	queuedAt time.Time,
) error {
	coll := mgm.Coll(&models.ApiLog{})
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":              status,
			"finishedAt":          now,
			"sandboxErrorType":    &errType,
			"sandboxErrorMessage": errMsg,
		},
	}

	result, err := coll.UpdateOne(
		ctx,
		bson.M{"jobId": jobID},
		update,
	)

	if err != nil {
		return err
	}

	// Safety check: job exists but apiLog missing
	if result.MatchedCount == 0 {
		return errors.New("api log not found for job")
	}

	return nil
}
