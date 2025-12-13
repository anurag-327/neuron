package conn

import (
	"sync"

	"github.com/docker/docker/client"
)

// DockerClient wraps the underlying Docker SDK client.
//
// This wrapper exists to allow future extension (helpers, metrics,
// retries) without leaking the raw Docker SDK throughout the codebase.
type DockerClient struct {
	*client.Client
}

var (
	dockerOnce   sync.Once
	dockerClient *DockerClient
	dockerErr    error
)

// GetDockerClient returns a singleton DockerClient instance.
//
// The client is initialized only once using environment configuration
// and API version negotiation. The returned client is safe for
// concurrent use and should be shared across the entire application.
func GetDockerClient() (*DockerClient, error) {
	dockerOnce.Do(func() {
		cli, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			dockerErr = err
			return
		}

		dockerClient = &DockerClient{
			Client: cli,
		}
	})

	return dockerClient, dockerErr
}
