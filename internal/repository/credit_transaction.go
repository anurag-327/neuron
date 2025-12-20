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

func CountCreditTransactionsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
) (int64, error) {
	coll := mgm.Coll(&models.CreditTransaction{})
	count, err := coll.CountDocuments(ctx, bson.M{"userId": userID})
	if err != nil {
		return 0, fmt.Errorf("failed to count credit transactions: %w", err)
	}
	return count, nil
}

// CreditStats represents aggregated credit statistics
type CreditStats struct {
	TotalCreditsEarned int64            `json:"totalCreditsEarned"`
	TotalCreditsSpent  int64            `json:"totalCreditsSpent"`
	NetCredits         int64            `json:"netCredits"`
	TotalTransactions  int64            `json:"totalTransactions"`
	CreditTransactions int64            `json:"creditTransactions"`
	DebitTransactions  int64            `json:"debitTransactions"`
	ByReason           map[string]int64 `json:"byReason"`
}

// GetCreditStatsByUserID aggregates credit statistics for a user within a date range
func GetCreditStatsByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	startDate, endDate primitive.DateTime,
) (*CreditStats, error) {

	coll := mgm.Coll(&models.CreditTransaction{})

	// Build match filter
	matchFilter := bson.M{"userId": userID}
	if !startDate.Time().IsZero() || !endDate.Time().IsZero() {
		dateFilter := bson.M{}
		if !startDate.Time().IsZero() {
			dateFilter["$gte"] = startDate
		}
		if !endDate.Time().IsZero() {
			dateFilter["$lte"] = endDate
		}
		matchFilter["created_at"] = dateFilter
	}

	// Aggregation pipeline
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"totalCreditsEarned": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$type", models.CreditTxnTypeCredit}},
						"$amount",
						0,
					},
				},
			},
			"totalCreditsSpent": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$type", models.CreditTxnTypeDebit}},
						"$amount",
						0,
					},
				},
			},
			"totalTransactions": bson.M{"$sum": 1},
			"creditTransactions": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$type", models.CreditTxnTypeCredit}},
						1,
						0,
					},
				},
			},
			"debitTransactions": bson.M{
				"$sum": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$type", models.CreditTxnTypeDebit}},
						1,
						0,
					},
				},
			},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate credit stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []struct {
		TotalCreditsEarned int64 `bson:"totalCreditsEarned"`
		TotalCreditsSpent  int64 `bson:"totalCreditsSpent"`
		TotalTransactions  int64 `bson:"totalTransactions"`
		CreditTransactions int64 `bson:"creditTransactions"`
		DebitTransactions  int64 `bson:"debitTransactions"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode credit stats: %w", err)
	}

	stats := &CreditStats{
		ByReason: make(map[string]int64),
	}

	if len(results) > 0 {
		stats.TotalCreditsEarned = results[0].TotalCreditsEarned
		stats.TotalCreditsSpent = results[0].TotalCreditsSpent
		stats.NetCredits = results[0].TotalCreditsEarned - results[0].TotalCreditsSpent
		stats.TotalTransactions = results[0].TotalTransactions
		stats.CreditTransactions = results[0].CreditTransactions
		stats.DebitTransactions = results[0].DebitTransactions
	}

	// Get breakdown by reason
	reasonPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$reason",
			"total": bson.M{"$sum": "$amount"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	reasonCursor, err := coll.Aggregate(ctx, reasonPipeline)
	if err != nil {
		return stats, nil // Return stats without reason breakdown
	}
	defer reasonCursor.Close(ctx)

	var reasonResults []struct {
		Reason string `bson:"_id"`
		Total  int64  `bson:"total"`
		Count  int64  `bson:"count"`
	}

	if err := reasonCursor.All(ctx, &reasonResults); err == nil {
		for _, r := range reasonResults {
			stats.ByReason[r.Reason] = r.Total
		}
	}

	return stats, nil
}
