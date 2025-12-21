package sandbox

import (
	"context"

	"github.com/anurag-327/neuron/pkg/sandbox/docker"
)

type Runner interface {
	Run(ctx context.Context, containerID, basePath, code, input, language string) docker.RunResult
	Health() error
}
