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

type RequestStatus string

const (
	RequestAccepted RequestStatus = "accepted"
	RequestRejected RequestStatus = "rejected"
	RequestError    RequestStatus = "error"
)

// ApiStats represents aggregated API request statistics
type ApiStats struct {
	TotalRequests   int64            `json:"totalRequests"`
	SuccessRequests int64            `json:"successRequests"`
	FailedRequests  int64            `json:"failedRequests"`
	ErrorRequests   int64            `json:"errorRequests"`
	ByEndpoint      map[string]int64 `json:"byEndpoint"`
	ByStatus        map[string]int64 `json:"byStatus"`
	SuccessRate     float64          `json:"successRate"`
}

type ApiLog struct {
	mgm.DefaultModel `bson:",inline"`

	UserID primitive.ObjectID  `bson:"userId" json:"userId"`
	JobID  *primitive.ObjectID `bson:"jobId,omitempty" json:"jobId,omitempty"`

	Endpoint      string        `bson:"endpoint" json:"endpoint"`
	Method        string        `bson:"method" json:"method"`
	ResponseCode  int64         `bson:"requestCode" json:"requestCode"`
	RequestStatus RequestStatus `bson:"requestStatus" json:"requestStatus"`
	RequestBody   string        `bson:"requestBody" json:"requestBody"`
	ErrorMessage  string        `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
	Status        RunStatus     `bson:"status" json:"status"`

	SandboxErrorType    *SandboxError `bson:"sandboxErrorType,omitempty" json:"sandboxErrorType,omitempty"`
	SandboxErrorMessage string        `bson:"sandboxErrorMessage,omitempty" json:"sandboxErrorMessage,omitempty"`

	StartedAt  time.Time `bson:"startedAt,omitempty" json:"started_at,omitempty"`
	FinishedAt time.Time `bson:"finishedAt,omitempty" json:"finished_at,omitempty"`
	QueuedAt   time.Time `bson:"queuedAt,omitempty" json:"queued_at,omitempty"`
}

func CreateApiLogIndexes() error {
	coll := mgm.Coll(&ApiLog{})

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().
				SetName("user_apilog_createdAt_idx"),
		},
		{
			Keys: bson.D{
				{Key: "jobId", Value: 1},
			},
			Options: options.Index().
				SetName("job_apilog_idx").
				SetSparse(true),
		},
	}

	_, err := coll.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return err
	}

	log.Println("ApiLog indexes created successfully")
	return nil
}
