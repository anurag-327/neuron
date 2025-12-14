package sandbox

import (
	"context"

	"github.com/anurag-327/neuron/internal/models"
)

type Runner interface {
	Run(ctx context.Context, basePath, code, input, language string) (string, string, models.SandboxError, string)
}
