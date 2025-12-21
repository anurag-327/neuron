package status

import (
	"net/http"
	"time"

	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetStatus(c *gin.Context) {
	// 1. Try to get status from DB
	storedStatus, err := repository.GetSystemStatus(c.Request.Context())
	if err == nil && storedStatus != nil {
		// Check if within 3 minutes
		if time.Since(storedStatus.UpdatedAt) < 3*time.Minute {
			response := gin.H{
				"publisher":  storedStatus.Publisher,
				"subscriber": storedStatus.Subscriber,
				"runner":     storedStatus.Runner,
				"updated_at": storedStatus.UpdatedAt,
			}

			c.JSON(http.StatusOK, response)
			return
		}
	}

	// 2. Recalculate
	pubHealth := factory.GetPublisherHealth()
	subHealth := factory.GetSubscriberHealth()
	runnerHealth := factory.GetRunnerHealth()

	newStatus := &models.SystemStatus{
		UpdatedAt:       time.Now(),
		Publisher:       models.StatusUp,
		Subscriber:      models.StatusUp,
		Runner:          models.StatusUp,
		PublisherError:  "",
		SubscriberError: "",
		RunnerError:     "",
	}

	if pubHealth != nil {
		newStatus.Publisher = models.StatusDown
		newStatus.PublisherError = pubHealth.Error()
	}
	if subHealth != nil {
		newStatus.Subscriber = models.StatusDown
		newStatus.SubscriberError = subHealth.Error()
	}
	if runnerHealth != nil {
		newStatus.Runner = models.StatusDown
		newStatus.RunnerError = runnerHealth.Error()
	}

	// 3. Save to DB
	_ = repository.UpsertSystemStatus(c.Request.Context(), newStatus)

	// 4. Respond
	response := gin.H{
		"publisher":  newStatus.Publisher,
		"subscriber": newStatus.Subscriber,
		"runner":     newStatus.Runner,
		"updated_at": newStatus.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}
