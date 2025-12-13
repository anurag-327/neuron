package pool

import (
	"sync"

	"github.com/docker/docker/client"
)

var (
	dockerOnce   sync.Once
	dockerClient *client.Client
	dockerErr    error
)

// GetDockerClient returns a singleton Docker client instance.
//
// The client is initialized only once using environment configuration
// and API version negotiation. The Docker client is safe for concurrent use
// and should be shared across all container pools.
func GetDockerClient() (*client.Client, error) {
	dockerOnce.Do(func() {
		dockerClient, dockerErr = client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
	})
	return dockerClient, dockerErr
}
