package repository

import (
	"context"

	"github.com/anurag-327/neuron/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const GlobalSystemStatusKey = "global_system_status"

func GetSystemStatus(ctx context.Context) (*models.SystemStatus, error) {
	var status models.SystemStatus
	err := models.SystemStatusCollection.FindOne(ctx, bson.M{"key": GlobalSystemStatusKey}).Decode(&status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func UpsertSystemStatus(ctx context.Context, status *models.SystemStatus) error {
	status.Key = GlobalSystemStatusKey
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"key": GlobalSystemStatusKey}
	update := bson.M{"$set": status}

	_, err := models.SystemStatusCollection.UpdateOne(ctx, filter, update, opts)
	return err
}
