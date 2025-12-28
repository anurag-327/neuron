package runnerHandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/internal/dto"
	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/registry"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/anurag-327/neuron/internal/services"
	"github.com/anurag-327/neuron/internal/util"
	"github.com/anurag-327/neuron/internal/util/response"
	"github.com/gin-gonic/gin"
)

func SubmitCodeHandler(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	apiLog := &models.ApiLog{
		UserID:        user.ID,
		JobID:         nil,
		Endpoint:      c.Request.URL.String(),
		Method:        c.Request.Method,
		ResponseCode:  http.StatusOK,
		RequestStatus: "success",
		Status:        "running",
		ErrorMessage:  "",
	}

	var body dto.SubmitCodeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apiLog.ResponseCode = http.StatusBadRequest
		apiLog.RequestStatus = "failed"
		apiLog.ErrorMessage = err.Error()
		apiLog.Status = "failed"
		_, _ = repository.SaveApiLog(ctx, apiLog)
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Convert body to JSON string for logging
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		bodyJSON = []byte("{}")
	}

	apiLog.RequestBody = string(bodyJSON)

	// 1 Language supported?
	langCfg, ok := registry.LanguageRegistry[body.Language]
	if !ok {
		apiLog.ResponseCode = http.StatusBadRequest
		apiLog.RequestStatus = "failed"
		apiLog.ErrorMessage = "language not supported"
		apiLog.Status = "failed"
		_, _ = repository.SaveApiLog(ctx, apiLog)
		response.Error(c, http.StatusBadRequest, "language not supported")
		return
	}

	// 2 Code validation
	if langCfg.Validator != nil {
		if err := langCfg.Validator(body.Code); err != nil {
			apiLog.ResponseCode = http.StatusBadRequest
			apiLog.RequestStatus = "failed"
			apiLog.ErrorMessage = err.Error()
			apiLog.Status = "failed"
			_, _ = repository.SaveApiLog(ctx, apiLog)
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	// 3 Credit check
	if err := services.AssertCanSubmit(ctx, user.ID); err != nil {
		if errors.Is(err, repository.ErrInsufficientCredits) {
			apiLog.ResponseCode = http.StatusPaymentRequired
			apiLog.RequestStatus = "failed"
			apiLog.Status = "failed"
			apiLog.ErrorMessage = "insufficient credits"
			_, _ = repository.SaveApiLog(ctx, apiLog)
			response.Error(c, http.StatusPaymentRequired, "insufficient credits")
			return
		}
		apiLog.ResponseCode = http.StatusInternalServerError
		apiLog.RequestStatus = "failed"
		apiLog.ErrorMessage = err.Error()
		_, _ = repository.SaveApiLog(ctx, apiLog)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 4 Create job
	job, err := services.CreateSubmission(ctx, user, body)
	if err != nil {
		apiLog.ResponseCode = http.StatusInternalServerError
		apiLog.RequestStatus = "failed"
		apiLog.Status = "failed"
		apiLog.ErrorMessage = err.Error()
		_, _ = repository.SaveApiLog(ctx, apiLog)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 5 Publish job
	jobBytes, _ := json.Marshal(job)
	p, err := factory.GetPublisher()
	if err != nil {
		apiLog.ResponseCode = http.StatusInternalServerError
		apiLog.RequestStatus = "failed"
		apiLog.Status = "failed"
		apiLog.ErrorMessage = "publisher unavailable"
		_, _ = repository.SaveApiLog(ctx, apiLog)
		_ = repository.DeleteJob(ctx, job)
		response.Error(c, http.StatusInternalServerError, "publisher unavailable")
		return
	}

	if err := p.Publish(config.ExecutionTasksTopic, job.Language, jobBytes); err != nil {
		apiLog.ResponseCode = http.StatusInternalServerError
		apiLog.RequestStatus = "failed"
		apiLog.Status = "failed"
		apiLog.ErrorMessage = err.Error()
		_, _ = repository.SaveApiLog(ctx, apiLog)
		_ = repository.DeleteJob(ctx, job)
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 6 Update api log
	apiLog.ResponseCode = http.StatusOK
	apiLog.RequestStatus = "success"
	apiLog.Status = "success"
	apiLog.ErrorMessage = ""
	apiLog.JobID = &job.ID
	_, _ = repository.SaveApiLog(ctx, apiLog)

	response.Success(
		c,
		http.StatusOK,
		"job queued successfully",
		gin.H{"jobId": job.ID, "status": job.Status},
	)
}

func GetJobStatusHandler(c *gin.Context) {

	ctx := c.Request.Context()

	user, err := util.GetUserFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	jobID := c.Param("jobId")
	if jobID == "" {
		response.Error(c, http.StatusBadRequest, "jobId is required")
		return
	}

	objID, err := util.IsValidObjectID(jobID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid Job ID")
		return
	}

	job, err := repository.GetJobByIDAndUserID(ctx, objID, user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Return minimal info for queued or running jobs
	if job.Status == models.StatusQueued || job.Status == models.StatusRunning {
		response.Success(c, http.StatusOK, "status fetched successfully", gin.H{
			"status": job.Status,
			"jobId":  job.ID,
		})
		return
	}

	executionTime := job.FinishedAt.Sub(job.StartedAt)
	queueTime := job.StartedAt.Sub(job.QueuedAt)
	totalTime := job.FinishedAt.Sub(job.QueuedAt)

	util.SetCreditsLeftHeader(c, user.Credits)
	response.Success(c, http.StatusOK, "job result fetched successfully", gin.H{
		"jobId":               job.ID,
		"status":              job.Status,
		"stdout":              job.Stdout,
		"stderr":              job.Stderr,
		"sandboxErrorType":    job.SandboxErrorType,
		"sandboxErrorMessage": job.SandboxErrorMessage,
		"language":            job.Language,
		"exitCode":            job.ExitCode,

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
