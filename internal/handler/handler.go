package handler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/anurag-327/neuron/internal/dto"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func SubmitCodeHandler(c *gin.Context) {
	now := time.Now()
	ctx := c.Request.Context()
	var body dto.SubmitCodeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	if body.Language == "cpp" {
		// Checks on code blocks
		if err := util.ValidateAndSanitizeCpp(body.Code); err != nil {
			response.Error(c, 400, err.Error())
			return
		}
	}

	// Save the job in db
	job := &models.Job{
		Language: body.Language,
		Code:     body.Code,
		Input:    *body.Input,
		Status:   models.StatusQueued,
		QueuedAt: now,
	}
	job, err := repository.SaveJob(ctx, job)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	// Publish it to kafka queue
	jobBytes, err := json.Marshal(job)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	p, err := factory.GetPublisher()
	if err != nil {
		response.Error(c, 500, "Internal Server Error")
	}
	if err := p.Publish("code-jobs", job.Language, jobBytes); err != nil {
		_ = repository.DeleteJob(ctx, job)
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, 200, "job queued successfully", gin.H{"jobId": job.ID, "status": job.Status})
}
func GetJobStatusHandler(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		response.Error(c, 400, "jobId is required")
		return
	}

	_, err := util.IsValidObjectID(jobID)
	if err != nil {
		response.Error(c, 401, "Invalid Job ID")
		return
	}

	ctx := c.Request.Context()

	job, err := repository.GetJobByID(ctx, jobID)
	fmt.Println(job, err)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	// Return minimal info for queued or running jobs
	if job.Status == models.StatusQueued || job.Status == models.StatusRunning {
		response.Success(c, 200, "status fetched successfully", gin.H{
			"status": job.Status,
			"jobId":  job.ID,
		})
		return
	}

	executionTime := job.FinishedAt.Sub(job.StartedAt)
	queueTime := job.StartedAt.Sub(job.QueuedAt)
	totalTime := job.FinishedAt.Sub(job.QueuedAt)

	response.Success(c, 200, "job result fetched successfully", gin.H{
		"jobId":               job.ID,
		"status":              job.Status,
		"stdout":              job.Stdout,
		"stderr":              job.Stderr,
		"sandboxErrorType":    job.SandboxErrorType,
		"sandboxErrorMessage": job.SandboxErrorMessage,
		"language":            job.Language,

		// timestamps
		"queuedAt":   job.QueuedAt,
		"startedAt":  job.StartedAt,
		"finishedAt": job.FinishedAt,

		// time statistics
		"executionTimeMs": executionTime.Milliseconds(),
		"queueTimeMs":     queueTime.Milliseconds(),
		"totalTimeMs":     totalTime.Milliseconds(),
	})

}
