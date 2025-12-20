package models

import (
	"context"
	"fmt"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CreditTransactionType string
type CreditTransactionReason string

const (
	// Transaction types
	CreditTxnTypeCredit CreditTransactionType = "credit"
	CreditTxnTypeDebit  CreditTransactionType = "debit"

	// Reasons
	// Credits IN
	CreditReasonSignupBonus CreditTransactionReason = "signup_bonus"
	CreditReasonGrant       CreditTransactionReason = "grant"
	CreditReasonPurchase    CreditTransactionReason = "purchase"
	CreditReasonRefund      CreditTransactionReason = "refund"
	CreditReasonDailyBonus  CreditTransactionReason = "daily_bonus"

	// Credits OUT
	CreditReasonSubmission CreditTransactionReason = "submission"
	CreditReasonRerun      CreditTransactionReason = "rerun"
)

type CreditTransaction struct {
	mgm.DefaultModel `bson:",inline"`

	UserID       primitive.ObjectID      `bson:"userId" json:"userId"`
	Type         CreditTransactionType   `bson:"type" json:"type"`
	Amount       int64                   `bson:"amount" json:"amount"`
	Reason       CreditTransactionReason `bson:"reason" json:"reason"`
	ReferenceID  *primitive.ObjectID     `bson:"referenceId,omitempty" json:"referenceId,omitempty"`
	BalanceAfter int64                   `bson:"balanceAfter" json:"balanceAfter"`
	Metadata     bson.M                  `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

func (c *CreditTransaction) CollectionName() string {
	return "credit_transactions"
}

func CreateCreditTransactionIndexes() error {
	coll := mgm.Coll(&CreditTransaction{})

	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"userId":     1,
				"created_at": -1,
			},
			Options: options.Index().
				SetName("user_time_idx"),
		},
		{
			Keys: bson.M{
				"referenceId": 1,
			},
			Options: options.Index().
				SetName("reference_idx").
				SetSparse(true),
		},
	}

	_, err := coll.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	fmt.Println("Credit transaction indexes created successfully")
	return nil
}
