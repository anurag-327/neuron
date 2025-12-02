package factory

import (
	"log"
	"sync"

	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/anurag-327/neuron/pkg/sandbox/docker"
)

var (
	runnerInstance     sandbox.Runner
	onceRunnerInstance sync.Once
)

func GetClient() sandbox.Runner {
	onceRunnerInstance.Do(func() {
		cli, err := docker.GetDockerClient()
		if err != nil {
			log.Fatalf("Failed to init docker client: %v", err)
		}
		runnerInstance = cli
	})
	return runnerInstance
}
