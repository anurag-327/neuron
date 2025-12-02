package sandboxUtil

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/pkg/sandbox"
)

func failJob(ctx context.Context, job *models.Job, errType sandbox.SandboxError, message string) error {
	job.Status = models.StatusFailed
	job.FinishedAt = time.Now()
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

func ExecuteCode(jobBytes []byte) error {
	var job models.Job
	ctx := context.Background()

	// Parse job
	if err := json.Unmarshal(jobBytes, &job); err != nil {
		failJob(ctx, &job, sandbox.ErrInternalError, "Malformed job payload")
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	// Set job → running
	job.Status = models.StatusRunning
	job.StartedAt = time.Now()

	if _, err := repository.SaveJob(ctx, &job); err != nil {
		failJob(ctx, &job, sandbox.ErrInternalError, "Failed to update running state")
		return fmt.Errorf("cannot update job state: %w", err)
	}

	// Init runner
	r := factory.GetClient()

	// Execute code
	basePath := fmt.Sprintf("/tmp/runner/job_%s", job.ID.Hex())
	stdout, stderr, errType, errMsg := r.Run(ctx, basePath, job.Code, job.Input, job.Language)

	// Update completion
	job.FinishedAt = time.Now()
	job.Stdout = stdout
	job.Stderr = stderr
	job.SandboxErrorType = nil
	if errType != "" {
		job.SandboxErrorType = &errType
	}
	job.SandboxErrorMessage = errMsg

	switch errType {
	case sandbox.ErrSandboxError, sandbox.ErrInternalError:
		job.Status = models.StatusFailed
	default:
		job.Status = models.StatusSuccess
	}

	if _, err := repository.SaveJob(ctx, &job); err != nil {
		failJob(ctx, &job, sandbox.ErrInternalError, "Failed to write final job state")
		return err
	}

	return nil
}
