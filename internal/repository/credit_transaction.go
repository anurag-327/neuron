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

var ErrCreditTransactionNotFound = errors.New("credit transaction not found")
var ErrInsufficientCredits = errors.New("insufficient credits")

func CreateCreditTransaction(
	ctx context.Context,
	txn *models.CreditTransaction,
) (*models.CreditTransaction, error) {

	if txn == nil {
		return nil, fmt.Errorf("credit transaction cannot be nil")
	}

	coll := mgm.Coll(txn)

	if err := coll.CreateWithCtx(ctx, txn); err != nil {
		return nil, fmt.Errorf("failed to create credit transaction: %w", err)
	}

	return txn, nil
}

// Get credit transactions for a user (latest first)
func GetCreditTransactionsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	page int64,
	limit int64,
) ([]models.CreditTransaction, error) {

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

	coll := mgm.Coll(&models.CreditTransaction{})

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
	var txns []models.CreditTransaction
	if err := cursor.All(ctx, &txns); err != nil {
		return nil, fmt.Errorf("failed to decode credit transactions: %w", err)
	}

	return txns, nil
}

// Get single transaction by ID (read-only)
func GetCreditTransactionByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*models.CreditTransaction, error) {

	txn := &models.CreditTransaction{}
	coll := mgm.Coll(txn)

	err := coll.FindByIDWithCtx(ctx, id, txn)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrCreditTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get credit transaction by id: %w", err)
	}

	return txn, nil
}
