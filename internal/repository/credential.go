package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrCredentialNotFound = errors.New("credential not found")

// CreateCredential creates a new API key for the user
func CreateCredential(ctx context.Context, cred *models.Credential) (*models.Credential, error) {
	coll := mgm.Coll(cred)
	if err := coll.CreateWithCtx(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}
	return cred, nil
}

// GetCredentialByUserID fetches the credential for a user
func GetCredentialByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Credential, error) {
	cred := &models.Credential{}
	coll := mgm.Coll(cred)

	err := coll.First(bson.M{"userId": userID}, cred)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return cred, nil
}

// GetCredentialByKey fetches a credential by its API key string
func GetCredentialByKey(ctx context.Context, key string) (*models.Credential, error) {
	cred := &models.Credential{}
	coll := mgm.Coll(cred)

	err := coll.First(bson.M{"key": key}, cred)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get credential by key: %w", err)
	}
	return cred, nil
}

// UpdateCredential updates the credential (e.g., LastUsedAt, IsActive)
func UpdateCredential(ctx context.Context, cred *models.Credential) error {
	coll := mgm.Coll(cred)
	return coll.UpdateWithCtx(ctx, cred)
}

// DeleteCredential deletes the credential (revoke)
func DeleteCredential(ctx context.Context, cred *models.Credential) error {
	coll := mgm.Coll(cred)
	return coll.DeleteWithCtx(ctx, cred)
}

// UpdateLastUsedAt updates the LastUsedAt timestamp
func UpdateCredentialLastUsedAt(ctx context.Context, id primitive.ObjectID) error {
	// This is an optimized update that doesn't require fetching the whole doc first if we had ID,
	// but using filtered update is safer if we only have ID.
	coll := mgm.Coll(&models.Credential{})
	now := time.Now()
	_, err := coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"lastUsedAt": now},
	})
	return err
}
