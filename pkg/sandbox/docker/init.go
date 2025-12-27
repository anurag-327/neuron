package docker

import (
	"context"
	"log"

	"github.com/anurag-327/neuron/config"
	"github.com/anurag-327/neuron/conn"
	"github.com/anurag-327/neuron/pkg/sandbox/docker/pool"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// InitDockerPool registers and warms up all supported language pools.
//
// This function should be invoked once during application startup.
func InitDockerPool(ctx context.Context) error {
	log.Println("Initializing sandbox container pools...")

	// Clean up any orphaned containers from previous runs
	if err := cleanupOrphanedContainers(ctx); err != nil {
		log.Printf(" Warning: Failed to cleanup orphaned containers: %v", err)
	}

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

	log.Println("Container pools warmed and ready!")
	return nil
}

// cleanupOrphanedContainers removes all containers created by Neuron
// that are still running from previous worker instances
func cleanupOrphanedContainers(ctx context.Context) error {
	client, err := conn.GetDockerClient()
	if err != nil {
		return err
	}

	// Get all language images we use
	images := make(map[string]bool)
	for _, cfg := range config.DockerPools() {
		images[cfg.Image] = true
	}

	// List all containers (running and stopped) that match our images
	filterArgs := filters.NewArgs()
	for image := range images {
		filterArgs.Add("ancestor", image)
	}

	containers, err := client.ContainerList(ctx, container.ListOptions{
		All:     true, // Include stopped containers
		Filters: filterArgs,
	})
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		log.Println("✅ No orphaned containers found")
		return nil
	}

	log.Printf("🧹 Found %d orphaned containers, cleaning up...", len(containers))

	// Remove each container
	removed := 0
	for _, c := range containers {
		err := client.ContainerRemove(ctx, c.ID, container.RemoveOptions{
			Force: true, // Force remove even if running
		})
		if err != nil {
			log.Printf("Failed to remove container %s: %v", c.ID[:12], err)
		} else {
			removed++
			log.Printf("Removed orphaned container: %s (image: %s)", c.ID[:12], c.Image)
		}
	}

	log.Printf("Cleaned up %d/%d orphaned containers", removed, len(containers))
	return nil
}
