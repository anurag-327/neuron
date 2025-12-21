package factory

import (
	"log"
	"sync"

	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/pkg/sandbox"
	"github.com/anurag-327/neuron/pkg/sandbox/docker"
)

var (
	runnerInstance     sandbox.Runner
	onceRunnerInstance sync.Once
)

func GetRunner() sandbox.Runner {
	onceRunnerInstance.Do(func() {
		client, err := conn.GetDockerClient()
		if err != nil {
			log.Fatalf("Failed to init docker client: %v", err)
		}
		runnerInstance = docker.NewRunner(client)
	})
	return runnerInstance
}

func GetRunnerHealth() error {
	r := GetRunner()
	if r == nil {
		return nil // or error if it should be initialized
	}
	return r.Health()
}
