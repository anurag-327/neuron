package pool

import (
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/api/types/container"
)

// healthLoop periodically checks the health of idle containers.
//
// It runs as a background goroutine and executes at a fixed interval.
// Unhealthy containers are removed and replaced to keep the pool usable.
//
// NOTE:
//   - This loop only checks *idle* containers, never in-use ones.
//   - It runs until the process exits (should ideally be cancellable via context).
func (p *ContainerPool) healthLoop() {
	interval := p.cfg.HealthInterval
	if interval <= 0 {
		interval = 2 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		p.checkAll()
	}
}

// checkAll iterates over all currently idle containers and verifies their health.
//
// For each idle container:
//   - If healthy: it is returned to the idle pool.
//   - If unhealthy: it is forcefully removed and replaced with a new container.
//
// This method ensures that the idle pool contains only ready-to-use containers.
func (p *ContainerPool) checkAll() {
	// Snapshot the number of idle containers at this moment.
	// We only iterate over this many to avoid infinite loops.
	n := len(p.idle)

	healthy := 0
	for i := 0; i < n; i++ {
		// Take one container out of the idle pool
		id := <-p.idle

		if p.isHealthy(id) {
			// Container is healthy → return it to idle pool
			p.idle <- id
			healthy++
			continue
		} else {
			p.ReplaceContainer(id)
		}
	}

	ratio := float64(healthy) / float64(n)

	switch {
	case ratio >= 0.8:
		p.setHealth(PoolHealthy)
	case ratio >= 0.4:
		p.setHealth(PoolDegraded)
	default:
		p.setHealth(PoolUnhealthy)
	}
}

// isHealthy determines whether a container is healthy.
//
// It performs a lightweight Docker exec using the configured HealthCmd.
// If the exec can be successfully created, the container is considered healthy.
//
// NOTE:
//   - This only validates container *liveness*, not correctness of execution.
//   - Exit codes and command output are not currently inspected.
func (p *ContainerPool) isHealthy(id string) bool {
	stats, err := p.client.ContainerStatsOneShot(context.Background(), id)
	if err != nil {
		return false
	}
	defer stats.Body.Close()

	var v container.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return false
	}

	// 1. PID CHECK: Extremely strict for idle containers
	// Baseline is usually 1 (the sleep command). Anything > 3 is suspicious.
	if v.PidsStats.Current > 5 {
		return false
	}

	// 2. MEMORY CHECK: Check for 'bloat'
	// If an idle container is using 50MB+ just sitting there, Python didn't GC properly.
	if v.MemoryStats.Usage > (50 * 1024 * 1024) {
		return false
	}

	// 3. STATE CHECK
	inspect, err := p.client.ContainerInspect(context.Background(), id)
	return err == nil && inspect.State.Running
}

func (p *ContainerPool) setHealth(h PoolHealth) {
	p.healthMu.Lock()
	defer p.healthMu.Unlock()

	p.health = h
	p.lastHealthCheck = time.Now()
}

func (p *ContainerPool) Health() PoolHealth {
	p.healthMu.RLock()
	defer p.healthMu.RUnlock()
	return p.health
}
