package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrUserNotFound = errors.New("user not found")

func SaveUser(ctx context.Context, user *models.User) (*models.User, error) {
	coll := mgm.Coll(user)
	if user.ID.IsZero() {
		// New job → create
		if err := coll.CreateWithCtx(ctx, user); err != nil {
			return nil, err
		}
	} else {
		// Existing job → update
		if err := coll.UpdateWithCtx(ctx, user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	user := &models.User{}
	coll := mgm.Coll(user)
	err := coll.FindByIDWithCtx(ctx, id, user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	coll := mgm.Coll(user)

	err := coll.FindOne(ctx, bson.M{"email": email}).Decode(user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}
