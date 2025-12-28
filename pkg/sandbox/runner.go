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
	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/pkg/sandbox/docker"
	"github.com/anurag-327/neuron/pkg/sandbox/docker/pool"
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
	// 2) Initialize Docker runner
	// -----------------------------
	dC, dockerErr := conn.GetDockerClient()
	if dockerErr != nil {
		log.Println("[RUN] failed to get docker client:", dockerErr)
		failJob(ctx, &job, models.ErrSandboxError, "failed to initiate sandboxed container")
		return fmt.Errorf("failed to get docker client: %w", dockerErr)
	}
	r := docker.NewRunner(dC)

	// -----------------------------
	// 3) Acquire warm container
	// -----------------------------
	p := pool.Manager.GetPool(job.Language)
	if p == nil {
		log.Println("[RUN] no pool for language:", job.Language)
		failJob(ctx, &job, models.ErrInternalError, "unsupported language")
		return fmt.Errorf("no pool for language")
	}

	containerID, err := p.Get(ctx)
	if err != nil {
		log.Println("[RUN] pool exhausted:", err)
		failJob(ctx, &job, models.ErrInternalError, "failed to acquire a container")
		return fmt.Errorf("no available containers")
	}

	// NOTE:
	// DO NOT defer Put() here.
	// Container lifecycle depends on execution result.

	// -----------------------------
	// 4) Mark job as RUNNING
	// -----------------------------
	job.Status = models.StatusRunning
	job.StartedAt = time.Now()

	if _, err := repository.SaveJob(ctx, &job); err != nil {
		failJob(ctx, &job, models.ErrInternalError, "Failed to update running state")
		return fmt.Errorf("cannot update job state: %w", err)
	}

	// -----------------------------
	// 5) Execute user code
	// -----------------------------
	basePath := fmt.Sprintf("/tmp/runner/job_%s", job.ID.Hex())

	runResult := r.Run(
		ctx,
		containerID,
		basePath,
		job.Code,
		job.Input,
		job.Language,
	)

	// -----------------------------
	// 6) Handle container lifecycle
	// -----------------------------
	if runResult.ContainerDirty {
		log.Println("[POOL] destroying dirty container:", containerID)
		p.ReplaceContainer(containerID)
	} else {
		log.Println("[POOL] returning clean container:", containerID)
		p.Put(containerID)
	}

	// -----------------------------
	// 7) Persist execution result
	// -----------------------------
	job.FinishedAt = time.Now()
	job.Stdout = runResult.Stdout
	job.Stderr = runResult.Stderr

	job.SandboxErrorType = nil
	if runResult.ErrType != "" {
		job.SandboxErrorType = &runResult.ErrType
	}
	job.SandboxErrorMessage = runResult.ErrMsg
	job.ExitCode = runResult.ExitCode

	switch runResult.ErrType {
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
		executionTime := job.FinishedAt.Sub(job.StartedAt)
		queueTime := job.StartedAt.Sub(job.QueuedAt)
		totalTime := job.FinishedAt.Sub(job.QueuedAt)
		amount := config.GetCreditsForReason(models.CreditReasonSubmission)

		err = services.DeductCreditsAndLog(
			ctx,
			job.UserID,
			amount,
			models.CreditReasonSubmission,
			&job.ID,
			map[string]interface{}{
				"language":      job.Language,
				"executionTime": executionTime,
				"queueTime":     queueTime,
				"totalTime":     totalTime,
			},
		)

		if err != nil {
			log.Printf("credit deduction failed for job %s: %v", job.ID.Hex(), err)
		}
	}

	_ = services.UpdateApiLog(ctx, job.ID, job.Status, &runResult.ErrType, runResult.ErrMsg, job.StartedAt, job.FinishedAt, job.QueuedAt)
	return nil
}
