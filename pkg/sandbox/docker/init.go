package docker

import (
	"context"
	"log"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/pkg/sandbox/docker/pool"
)

// InitDockerPool registers and warms up all supported language pools.
//
// This function should be invoked once during application startup.
func InitDockerPool(ctx context.Context) error {
	log.Println("🔥 Initializing sandbox container pools...")

	for _, cfg := range config.DockerPools() {
		pool.Manager.Register(cfg.Language, pool.PoolConfig{
			Image:          cfg.Image,
			InitSize:       cfg.InitSize,
			MaxSize:        cfg.MaxSize,
			HealthCmd:      cfg.HealthCmd,
			HealthInterval: cfg.HealthInterval,
		})
	}

	if err := pool.Manager.InitAll(ctx); err != nil {
		return err
	}

	log.Println(" Container pools warmed and ready!")
	return nil
}
