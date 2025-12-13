package models

import (
	"time"

	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RunStatus string

const (
	StatusQueued  RunStatus = "queued"
	StatusRunning RunStatus = "running"
	StatusSuccess RunStatus = "success"
	StatusFailed  RunStatus = "failed"
)

type Job struct {
	mgm.DefaultModel `bson:",inline"`

	UserID primitive.ObjectID `bson:"userId" json:"userId"`

	Language            string                `bson:"language" json:"language"`
	Code                string                `bson:"code" json:"code"`
	Input               string                `bson:"input,omitempty" json:"input,omitempty"`
	Status              RunStatus             `bson:"status" json:"status"`
	Stdout              string                `bson:"stdout,omitempty" json:"stdout,omitempty"`
	Stderr              string                `bson:"stderr,omitempty" json:"stderr,omitempty"`
	SandboxErrorType    *sandbox.SandboxError `bson:"errorType,omitempty" json:"error_type,omitempty"`
	SandboxErrorMessage string                `bson:"errorMessage,omitempty" json:"error_message,omitempty"`
	ExitCode            *int                  `bson:"exitCode,omitempty" json:"exit_code,omitempty"`
	StartedAt           time.Time             `bson:"startedAt,omitempty" json:"started_at,omitempty"`
	FinishedAt          time.Time             `bson:"finishedAt,omitempty" json:"finished_at,omitempty"`
	QueuedAt            time.Time             `bson:"queuedAt,omitempty" json:"queued_at,omitempty"`

	User *User `bson:"-" json:"user,omitempty"`
}
