// Package sandbox contains helper utilities for executing sandbox jobs.
// It is responsible for:
//   - Parsing job payloads
//   - Acquiring containers from the pool
//   - Executing user code inside a sandbox
//   - Persisting job state transitions (running → success / failed)
package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/pkg/engine"
)

// failJob updates the job as FAILED and persists the failure state.
//
// This helper is used when execution cannot proceed or a fatal error occurs.
// It guarantees:
//   - job status is set to FAILED
//   - finish timestamp is recorded
//   - sandbox error type & message are stored
//
// NOTE:
// This function does NOT panic. It always attempts best-effort persistence.
func failJob(
	ctx context.Context,
	job *models.Job,
	errType models.SandboxError,
	message string,
) error {

	job.Status = models.StatusFailed
	job.FinishedAt = time.Now()

	// SandboxErrorType is nullable in DB
	job.SandboxErrorType = nil
	if errType != "" {
		job.SandboxErrorType = &errType
	}

	job.SandboxErrorMessage = message

	if _, err := repository.SaveJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job failure state: %w", err)
	}

	return nil
}

// ExecuteCode is the main entry point for sandbox execution.

// It is responsible for:
//   - Parsing job payloads
//   - Acquiring containers from the pool
//   - Executing user code inside a sandbox
//   - Persisting job state transitions (running → success / failed)

// Lifecycle:
//  1. Parse incoming job payload
//  2. Acquire a warm container from pool
//  3. Mark job as RUNNING
//  4. Execute user code inside sandbox
//  5. Persist stdout/stderr/results
//  6. Return container back to pool
//
// This function is intentionally synchronous:
// - Caller controls concurrency
// - Pool enforces execution limits
func ExecuteCode(jobBytes []byte) error {

	var job models.Job
	ctx := context.Background()

	// -----------------------------
	// 1) Parse job payload
	// -----------------------------
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		failJob(ctx, &job, models.ErrInternalError, "Malformed job payload")
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	// -----------------------------
	// 2) Mark job as RUNNING
	// -----------------------------
	job.Status = models.StatusRunning
	job.StartedAt = time.Now()

	if _, err := repository.SaveJob(ctx, &job); err != nil {
		failJob(ctx, &job, models.ErrInternalError, "Failed to update running state")
		return fmt.Errorf("cannot update job state: %w", err)
	}

	// -----------------------------
	// 3) Execute on Core Engine
	// -----------------------------
	client := engine.NewClient(config.NeuronCoreURL)

	req := engine.ExecuteRequest{
		Code:     job.Code,
		Input:    job.Input,
		Language: job.Language,
		Limit: engine.Limit{
			TimeMs:   3000,       // Default 3s
			MemoryKB: 256 * 1024, // Default 256MB
		},
	}

	runResult, err := client.Execute(ctx, req)
	if err != nil {
		log.Printf("[RUN] engine execution failed: %v", err)
		failJob(ctx, &job, models.ErrInternalError, "Execution engine failure")
		return fmt.Errorf("engine execution failed: %w", err)
	}

	// -----------------------------
	// 4) Persist execution result
	// -----------------------------
	job.FinishedAt = time.Now()
	job.Stdout = runResult.Stdout
	job.Stderr = runResult.Stderr
	job.ExitCode = runResult.ExitCode
	job.Metrics = models.Metrics{
		Compile: runResult.Metrics.Compile,
		Run:     runResult.Metrics.Run,
		Total:   runResult.Metrics.Total,
	}

	job.SandboxErrorType = nil
	if runResult.ErrType != "" {
		errType := models.SandboxError(runResult.ErrType)
		job.SandboxErrorType = &errType
	}
	job.SandboxErrorMessage = runResult.ErrMsg

	switch models.SandboxError(runResult.ErrType) {
	case models.ErrSandboxError, models.ErrInternalError:
		job.Status = models.StatusFailed
	default:
		job.Status = models.StatusSuccess
	}

	if _, err := repository.SaveJob(ctx, &job); err != nil {
		failJob(ctx, &job, models.ErrInternalError, "Failed to write final job state")
		return err
	}

	if runResult.ErrType == "" {
		queueTime := job.StartedAt.Sub(job.QueuedAt)
		amount := config.GetCreditsForReason(models.CreditReasonSubmission)

		err = services.DeductCreditsAndLog(
			ctx,
			job.UserID,
			amount,
			models.CreditReasonSubmission,
			&job.ID,
			map[string]interface{}{
				"language":    job.Language,
				"runTime":     job.Metrics.Run,
				"compileTime": job.Metrics.Compile,
				"queueTime":   queueTime.Milliseconds(),
				"totalTime":   job.Metrics.Total,
			},
		)

		if err != nil {
			log.Printf("credit deduction failed for job %s: %v", job.ID.Hex(), err)
		}
	}

	_ = services.UpdateApiLog(ctx, job.ID, job.Status, job.SandboxErrorType, runResult.ErrMsg, job.StartedAt, job.FinishedAt, job.QueuedAt)
	return nil
}
