package models

import (
	"time"

	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/kamva/mgm/v3"
)

type RunStatus string

const (
	StatusQueued  RunStatus = "queued"
	StatusRunning RunStatus = "running"
	StatusSuccess RunStatus = "success"
	StatusFailed  RunStatus = "failed"
)

type Job struct {
	mgm.DefaultModel    `bson:",inline"`
	Language            string                `bson:"language" json:"language"`
	Code                string                `bson:"code" json:"code"`
	Input               string                `bson:"input,omitempty" json:"input,omitempty"`
	Status              RunStatus             `bson:"status" json:"status"`
	Stdout              string                `bson:"stdout,omitempty" json:"stdout,omitempty"`
	Stderr              string                `bson:"stderr,omitempty" json:"stderr,omitempty"`
	SandboxErrorType    *sandbox.SandboxError `bson:"error_type,omitempty" json:"error_type,omitempty"`
	SandboxErrorMessage string                `bson:"error_message,omitempty" json:"error_message,omitempty"`
	ExitCode            *int                  `bson:"exit_code,omitempty" json:"exit_code,omitempty"`
	StartedAt           time.Time             `bson:"started_at,omitempty" json:"started_at,omitempty"`
	FinishedAt          time.Time             `bson:"finished_at,omitempty" json:"finished_at,omitempty"`
	QueuedAt            time.Time             `bson:"queued_at,omitempty" json:"queued_at,omitempty"`
}
