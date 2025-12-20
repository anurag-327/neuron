package repository

import (
	"context"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func GetApiLogsByUserID(ctx context.Context, userID primitive.ObjectID, page, limit int64) ([]models.ApiLog, error) {
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
		return nil, fmt.Errorf("failed to fetch credit transactions: %w", err)
	}

	defer cursor.Close(ctx)

	var logs []models.ApiLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode credit transactions: %w", err)
	}

	return logs, nil
}
