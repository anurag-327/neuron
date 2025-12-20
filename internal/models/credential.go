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

// Credential represents a user API key/credential
type Credential struct {
	mgm.DefaultModel `bson:",inline"`

	UserID     primitive.ObjectID `bson:"userId" json:"userId"`
	Key        string             `bson:"key" json:"key"`
	Env        string             `bson:"env" json:"env"`           // e.g., "live", "test"
	IsActive   bool               `bson:"isActive" json:"isActive"` // user can disable it
	LastUsedAt *time.Time         `bson:"lastUsedAt,omitempty" json:"lastUsedAt,omitempty"`
}

func CreateCredentialIndexes() error {
	coll := mgm.Coll(&Credential{})

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
			Options: options.Index().
				SetName("user_credential_unique_idx").
				SetUnique(true), // One user, one credential
		},
		{
			Keys: bson.D{
				{Key: "key", Value: 1},
			},
			Options: options.Index().
				SetName("credential_key_idx").
				SetUnique(true),
		},
	}

	_, err := coll.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	log.Println("Credential indexes created successfully")
	return nil
}
