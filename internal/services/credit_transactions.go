package services

import (
	"context"
	"fmt"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeductCreditsAndLog(
	ctx context.Context,
	userID primitive.ObjectID,
	amount int64,
	reason models.CreditTransactionReason,
	referenceID *primitive.ObjectID,
	metadata map[string]interface{},
) error {

	// 1. Deduct credits atomically and get balanceAfter
	balanceAfter, err := repository.DeductUserCredits(ctx, userID, amount)
	if err != nil {
		return err
	}

	// 2. Write ledger
	_, err = repository.CreateCreditTransaction(ctx, &models.CreditTransaction{
		UserID:       userID,
		Type:         models.CreditTxnTypeDebit,
		Amount:       amount,
		Reason:       reason,
		ReferenceID:  referenceID,
		BalanceAfter: balanceAfter,
		Metadata:     metadata,
	})

	if err != nil {
		// COMPENSATION: restore credits
		_, _ = repository.AddUserCredits(ctx, userID, amount)
		return fmt.Errorf("ledger write failed, credits restored: %w", err)
	}

	return nil
}

func CreditUserAndLog(
	ctx context.Context,
	userID primitive.ObjectID,
	amount int64,
	reason models.CreditTransactionReason,
	referenceID *primitive.ObjectID,
	metadata map[string]interface{},
) error {

	balanceAfter, err := repository.AddUserCredits(ctx, userID, amount)
	if err != nil {
		return err
	}

	_, err = repository.CreateCreditTransaction(ctx, &models.CreditTransaction{
		UserID:       userID,
		Type:         models.CreditTxnTypeCredit,
		Amount:       amount,
		Reason:       reason,
		ReferenceID:  referenceID,
		BalanceAfter: balanceAfter,
		Metadata:     metadata,
	})

	return err
}
