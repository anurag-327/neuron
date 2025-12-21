package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrCredentialAlreadyExists = errors.New("active credential already exists for this user")

// GenerateAPIKey generates a random API key string
// Format: nr_live_<32_char_hex>
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	randomPart := hex.EncodeToString(bytes)
	return fmt.Sprintf("nr_live_%s", randomPart), nil
}

// CreateCredential generates and saves a new credential for the user
func CreateCredential(ctx context.Context, userID primitive.ObjectID) (*models.Credential, error) {
	// Check if exists
	existing, err := repository.GetCredentialByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, ErrCredentialAlreadyExists
	}

	key, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	cred := &models.Credential{
		UserID:   userID,
		Key:      key,
		Env:      "live",
		IsActive: true,
	}

	return repository.CreateCredential(ctx, cred)
}

// GetCredential retrieves the user's credential
func GetCredential(ctx context.Context, userID primitive.ObjectID) (*models.Credential, error) {
	return repository.GetCredentialByUserID(ctx, userID)
}

// RevokeCredential deletes the user's credential
func RevokeCredential(ctx context.Context, userID primitive.ObjectID) error {
	cred, err := repository.GetCredentialByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrCredentialNotFound) {
			return nil // Already gone
		}
		return err
	}

	return repository.DeleteCredential(ctx, cred)
}

// ValidateAPIKey checks if the key exists and is active, returns the associated UserID
func ValidateAPIKey(ctx context.Context, key string) (*models.Credential, error) {
	cred, err := repository.GetCredentialByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	if !cred.IsActive {
		return nil, errors.New("credential is inactive")
	}

	// Async update last used
	go func(id primitive.ObjectID) {
		_ = repository.UpdateCredentialLastUsedAt(context.Background(), id)
	}(cred.ID)

	return cred, nil
}
