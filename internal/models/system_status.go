package models

import (
	"context"
	"log"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ComponentStatus string

const (
	StatusUp   ComponentStatus = "up"
	StatusDown ComponentStatus = "down"
)

type SystemStatus struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key       string             `bson:"key" json:"key"` // singleton key, e.g. "global_status"
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`

	Publisher       ComponentStatus `bson:"publisher" json:"publisher"`
	PublisherError  string          `bson:"publisherError,omitempty" json:"publisherError,omitempty"`
	Subscriber      ComponentStatus `bson:"subscriber" json:"subscriber"`
	SubscriberError string          `bson:"subscriberError,omitempty" json:"subscriberError,omitempty"`
	Runner          ComponentStatus `bson:"runner" json:"runner"`
	RunnerError     string          `bson:"runnerError,omitempty" json:"runnerError,omitempty"`
}

var SystemStatusCollection *mgm.Collection

func CreateSystemStatusIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Using mgm to get the collection, assuming connection is already established
	SystemStatusCollection = mgm.CollectionByName("system_status")

	// Ensure unique index on Key to enforce singleton
	_, err := SystemStatusCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Failed to create index for system_status: %v", err)
	}
}
