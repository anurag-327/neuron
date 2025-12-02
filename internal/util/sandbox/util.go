package sandboxUtil

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anurag-327/neuron/internal/factory"
	"github.com/anurag-327/neuron/internal/models"
)

func ExecuteCode(jobBytes []byte) error {
	var job models.Job
	err := json.Unmarshal(jobBytes, &job)
	if err != nil {
		return err
	}
	ctx := context.Background()
	r := factory.GetClient()

	basePath := fmt.Sprintf("/tmp/runner/job_%s", job.ID.Hex())

	stdout, stderr, errType, errMsg := r.Run(ctx, basePath, job.Code, job.Input, job.Language)

	fmt.Println("Stdout", stdout)
	fmt.Println("Stderr:", stderr)
	fmt.Println("ErrType:", errType)
	fmt.Println("ErrorMessage", errMsg)

	return nil
}
