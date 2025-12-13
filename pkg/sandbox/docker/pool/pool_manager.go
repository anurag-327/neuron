package pool

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// PoolManager manages multiple container pools keyed by language/runtime.
//
// It is responsible for:
//   - registering pools
//   - warming them on startup
//   - providing lookup access
//   - graceful destruction during shutdown
//
// PoolManager is safe for concurrent use.
type PoolManager struct {
	mu sync.Mutex

	// pools maps language → container pool
	pools map[string]*ContainerPool
}

// Manager is the global singleton pool manager.
//
// This is initialized at process startup and shared across the platform.
var Manager = &PoolManager{
	pools: map[string]*ContainerPool{},
}

// Register registers a new container pool for a given language.
//
// If a pool with the same language already exists, it will be replaced.
// Pool creation errors are ignored here and should surface during InitAll.
func (pm *PoolManager) Register(language string, cfg PoolConfig) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, _ := NewPool(language, cfg)
	pm.pools[language] = pool
}

// InitAll pre-warms all registered container pools.
//
// It eagerly creates InitSize containers per pool to reduce
// cold-start latency during execution.
func (pm *PoolManager) InitAll(ctx context.Context) error {
	for lang, p := range pm.pools {
		log.Printf("🔥 Pre-warming container pool for %s...", lang)

		if err := p.WarmUp(ctx); err != nil {
			return err
		}
	}
	return nil
}

// GetPool returns the container pool for a given language.
//
// The caller must handle the case where the pool does not exist.
func (pm *PoolManager) GetPool(language string) *ContainerPool {
	return pm.pools[language]
}

// WarmUp eagerly creates InitSize containers and adds them to the idle pool.
//
// It also starts the background health-check loop.
func (p *ContainerPool) WarmUp(ctx context.Context) error {
	success := 0

	for i := 0; i < p.cfg.InitSize; i++ {
		id, err := p.newContainer(ctx)
		if err != nil {
			log.Printf(
				"Failed to warm container %d/%d for %s: %v",
				i+1, p.cfg.InitSize, p.lang, err,
			)
			continue
		}

		p.idle <- id
		p.mu.Lock()
		p.total++
		p.mu.Unlock()

		success++
	}

	if success == 0 {
		// Nothing usable → pool cannot function
		return fmt.Errorf("failed to warm any containers for %s", p.lang)
	}

	if success < p.cfg.InitSize {
		log.Printf(
			"Pool %s started in DEGRADED mode (%d/%d ready)",
			p.lang, success, p.cfg.InitSize,
		)
	}
	go p.healthLoop()
	return nil
}

// Get acquires a container from the pool.
//
// Behavior:
//  1. Try to reuse an idle container (fast path)
//  2. If capacity allows, create a new container (scale up)
//  3. Otherwise block until a container becomes available or context cancels
func (p *ContainerPool) Get(ctx context.Context) (string, error) {
	// Fast path: reuse idle container
	select {
	case id := <-p.idle:
		return id, nil
	default:
	}

	// Scale up if allowed
	p.mu.Lock()
	if p.total < p.cfg.MaxSize {
		log.Printf(" Scaling up pool for %s (%d → %d)",
			p.lang, p.total, p.total+1)

		id, err := p.newContainer(ctx)
		if err == nil {
			p.total++
			p.mu.Unlock()
			return id, nil
		}
	}
	p.mu.Unlock()

	// 3 Block until container available or context cancelled
	select {
	case id := <-p.idle:
		return id, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Put returns a container back to the pool.
//
// If the idle pool is full and the pool size exceeds InitSize,
// the container is destroyed to scale down.
func (p *ContainerPool) Put(id string) {
	// Attempt non-blocking return to idle pool
	select {
	case p.idle <- id:
		return
	default:
		// idle pool full → consider scale down
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Never scale below InitSize
	if p.total <= p.cfg.InitSize {
		// Blockingly return container to idle pool
		p.idle <- id
		return
	}

	// Scale down
	log.Printf("📉 Scaling down pool for %s (%d → %d)",
		p.lang, p.total, p.total-1)

	p.total--

	// Remove container asynchronously
	go func() {
		_ = p.client.ContainerRemove(
			context.Background(),
			id,
			container.RemoveOptions{Force: true},
		)
	}()
}

// newContainer creates and starts a new sandbox container.
//
// The container is started in an isolated network namespace and
// mounts a shared sandbox directory for code execution.
func (p *ContainerPool) newContainer(ctx context.Context) (string, error) {
	projectRoot, _ := os.Getwd()
	basePath := filepath.Join(projectRoot, "/tmp/runner")
	absPath, _ := filepath.Abs(basePath)

	resp, err := p.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:      p.cfg.Image,
			Cmd:        []string{"sleep", "infinity"},
			WorkingDir: "/app",
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: absPath, Target: "/sandbox"},
				{Type: mount.TypeTmpfs, Target: "/tmp"},
			},
			NetworkMode: "none",
		},
		nil, nil, "",
	)
	if err != nil {
		return "", err
	}

	if err := p.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

// replaceContainer removes an unhealthy container and attempts to
// replace it with a newly created one.
//
// This function is typically invoked by the background health-check
// loop when a container fails a health check. The replacement process
// is best-effort: failures are logged but do not block pool operation.
//
// NOTE:
//   - The removal is forced to ensure no leaked containers.
//   - If replacement fails, the pool may temporarily run with
//     reduced capacity until the next scaling or health cycle.
func (p *ContainerPool) replaceContainer(id string) {
	// Container is unhealthy → remove it from the system
	log.Printf("❌ Unhealthy container removed: %s", id)

	_ = p.client.ContainerRemove(
		context.Background(),
		id,
		container.RemoveOptions{Force: true},
	)

	// Attempt to spawn a replacement container to maintain pool capacity
	log.Printf("🔁 Spawning a replacement container")

	newID, err := p.newContainer(context.Background())
	if err != nil {
		// Replacement failure is non-fatal; capacity will be
		// restored by future scaling or health-check cycles
		log.Printf("⚠️ Failed to spawn replacement container: %v", err)
		return
	}

	// Return the new container to the idle pool
	p.idle <- newID
}

// DestroyAll gracefully destroys all pools and their containers.
//
// This should be called during application shutdown.
func (pm *PoolManager) DestroyAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	log.Println("Destroying all warm containers...")

	for lang, p := range pm.pools {
		log.Printf("Cleaning pool for %s", lang)
		p.Destroy()
	}

	log.Println(" All pools cleaned up")
}

// Destroy removes all containers managed by the pool.
func (p *ContainerPool) Destroy() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.idle)

	for id := range p.idle {
		log.Printf("Removing container %s", id)
		p.client.ContainerRemove(
			context.Background(),
			id,
			container.RemoveOptions{Force: true},
		)
	}
}
