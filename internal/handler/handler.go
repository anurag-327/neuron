package handler

import (
	"encoding/json"
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
	p := factory.GetPublisher()
	if err := p.Publish("code-jobs", job.Language, jobBytes); err != nil {
		_ = repository.DeleteJob(ctx, job)
		response.Error(c, 500, err.Error())
		return
	}

	// Return job id
	response.Success(c, 200, "job queued successfully", gin.H{"jobId": job.ID})
}

func GetJobStatusHandler(c *gin.Context) {
	jobID := c.Param("jobId")
	if jobID == "" {
		response.Error(c, 400, "jobId is required")
		return
	}
	ctx := c.Request.Context()
	// Get job from db
	job, err := repository.GetJobByID(ctx, jobID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	// Return status to user
	response.Success(c, 200, "status fetched successfully", gin.H{"status": job.Status})
}
