package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

// hashAPIKey creates a SHA-256 hash of the API key for secure storage
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CreateCredential generates and saves a new credential for the user
func CreateCredential(ctx context.Context, userID primitive.ObjectID) (*models.Credential, error) {
	// Check if exists
	existing, err := repository.GetCredentialByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, ErrCredentialAlreadyExists
	}

	plainKey, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	hashedKey := hashAPIKey(plainKey)

	cred := &models.Credential{
		UserID:   userID,
		Key:      hashedKey,
		Env:      "live",
		IsActive: true,
	}

	savedCred, err := repository.CreateCredential(ctx, cred)
	if err != nil {
		return nil, err
	}

	savedCred.Key = plainKey
	return savedCred, nil
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
			return nil
		}
		return err
	}

	return repository.DeleteCredential(ctx, cred)
}

// ValidateAPIKey checks if the key exists and is active, returns the associated UserID
func ValidateAPIKey(ctx context.Context, plainKey string) (*models.Credential, error) {
	// Hash the provided key to compare with stored hash
	hashedKey := hashAPIKey(plainKey)

	cred, err := repository.GetCredentialByKey(ctx, hashedKey)
	if err != nil {
		return nil, err
	}

	if !cred.IsActive {
		return nil, errors.New("credential is inactive")
	}

	go func(id primitive.ObjectID) {
		_ = repository.UpdateCredentialLastUsedAt(context.Background(), id)
	}(cred.ID)

	return cred, nil
}

// RegenerateCredential revokes the old credential and creates a new one
// Returns the new credential with the plain key
func RegenerateCredential(ctx context.Context, userID primitive.ObjectID) (*models.Credential, error) {
	// Check if credential exists
	existing, err := repository.GetCredentialByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Delete the old credential
	if err := repository.DeleteCredential(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to delete old credential: %w", err)
	}

	// Generate new plain key
	plainKey, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	hashedKey := hashAPIKey(plainKey)

	// Create new credential
	cred := &models.Credential{
		UserID:   userID,
		Key:      hashedKey,
		Env:      "live",
		IsActive: true,
	}

	savedCred, err := repository.CreateCredential(ctx, cred)
	if err != nil {
		return nil, err
	}

	// Return with plain key
	savedCred.Key = plainKey
	return savedCred, nil
}
