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
	"go.mongodb.org/mongo-driver/mongo/options"
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

func DeductUserCredits(
	ctx context.Context,
	userID primitive.ObjectID,
	amount int64,
) (int64, error) {

	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	coll := mgm.Coll(&models.User{})

	var updatedUser models.User
	err := coll.FindOneAndUpdate(
		ctx,
		bson.M{
			"_id":     userID,
			"credits": bson.M{"$gte": amount},
		},
		bson.M{
			"$inc": bson.M{"credits": -amount},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, ErrInsufficientCredits
		}
		return 0, err
	}

	return updatedUser.Credits, nil
}

func HasSufficientCredits(
	ctx context.Context,
	userID primitive.ObjectID,
	requiredCredits int64,
) error {

	if requiredCredits <= 0 {
		return errors.New("required credits must be greater than zero")
	}

	coll := mgm.Coll(&models.User{})

	count, err := coll.CountDocuments(
		ctx,
		bson.M{
			"_id":     userID,
			"credits": bson.M{"$gte": requiredCredits},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to check user credits: %w", err)
	}

	if count == 0 {
		return ErrInsufficientCredits
	}

	return nil
}

func AddUserCredits(
	ctx context.Context,
	userID primitive.ObjectID,
	amount int64,
) (int64, error) {

	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	coll := mgm.Coll(&models.User{})

	var updatedUser models.User
	err := coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$inc": bson.M{"credits": amount}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedUser)

	if err != nil {
		return 0, err
	}

	return updatedUser.Credits, nil
}
