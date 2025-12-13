package config

import "time"

// DockerPoolConfig defines configuration for a single language pool.
type DockerPoolConfig struct {
	Language       string
	Image          string
	InitSize       int
	MaxSize        int
	HealthCmd      []string
	HealthInterval time.Duration
}

// DockerPools returns the list of container pool configurations.
//
// This is the single source of truth for all supported
// runtimes and their pool behavior.
func DockerPools() []DockerPoolConfig {
	return []DockerPoolConfig{
		{
			Language:       "cpp",
			Image:          "gcc:latest",
			InitSize:       1,
			MaxSize:        2,
			HealthCmd:      []string{"echo", "ok"},
			HealthInterval: 40 * time.Second,
		},
		{
			Language:       "python",
			Image:          "python:3.12-alpine",
			InitSize:       1,
			MaxSize:        2,
			HealthCmd:      []string{"python3", "-c", "print('ok')"},
			HealthInterval: 20 * time.Second,
		},
		{
			Language:       "java",
			Image:          "eclipse-temurin:21-jdk-alpine",
			InitSize:       1,
			MaxSize:        2,
			HealthCmd:      []string{"java", "-version"},
			HealthInterval: 20 * time.Second,
		},
		{
			Language:       "js",
			Image:          "node:22-alpine",
			InitSize:       1,
			MaxSize:        2,
			HealthCmd:      []string{"node", "-version"},
			HealthInterval: 20 * time.Second,
		},
	}
}
