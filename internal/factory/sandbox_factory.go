package factory

import (
	"log"
	"sync"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/pkg/sandbox"
)

var (
	runnerInstance     sandbox.Runner
	onceRunnerInstance sync.Once
)

func GetRunner() sandbox.Runner {
	onceRunnerInstance.Do(func() {
		_, err := conn.GetDockerClient()
		if err != nil {
			log.Fatalf("Failed to init docker client: %v", err)
		}

	})
	return runnerInstance
}
