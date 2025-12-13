package pool

import (
	"sync"
	"time"

	"github.com/anurag-327/neuron/conn"
)

// PoolHealth represents the aggregated health state of a container pool.
//
// The health state is derived from the availability and condition
// of idle containers and is used to influence scheduling and
// traffic routing decisions.
type PoolHealth int

const (
	// PoolUnknown indicates that the pool health has not yet been evaluated.
	// This is typically the initial state during startup.
	PoolUnknown PoolHealth = iota

	// PoolHealthy indicates that the pool is fully operational and
	// has sufficient healthy containers to handle expected load.
	PoolHealthy

	// PoolDegraded indicates that the pool is partially operational.
	// Some containers may be unhealthy, but the pool can still
	// serve requests with reduced capacity.
	PoolDegraded

	// PoolUnhealthy indicates that the pool is not operational.
	// Requests should not be routed to this pool.
	PoolUnhealthy
)

// PoolConfig defines configuration options for a container pool.
//
// It controls which Docker image is used, how many containers are
// pre-warmed on startup, the maximum allowed containers, and
// the health check command used to verify container readiness.
type PoolConfig struct {
	// Image is the Docker image used to create containers
	Image string

	// InitSize is the number of containers created eagerly
	// when the pool is initialized.
	InitSize int

	// MaxSize is the hard upper limit on the total number
	// of containers (idle + in-use) managed by the pool.
	MaxSize int

	// HealthCmd is an optional command executed inside a container
	// to verify it is healthy and ready to accept work.
	// Example: []string{"python", "--version"}
	HealthCmd []string

	// HealthInterval defines how often health checks are performed
	// on idle containers in the pool.
	//
	// If set to zero or a negative value, a sensible default
	// (e.g., 2 minutes) will be used.
	HealthInterval time.Duration
}

// ContainerPool manages a pool of reusable Docker containers
// for a specific language/runtime.
//
// The pool maintains a set of idle containers that can be
// borrowed and returned to reduce container startup latency.
// It is safe for concurrent use.
type ContainerPool struct {
	// lang identifies the language/runtime this pool serves
	lang string

	// cfg holds the pool configuration.
	cfg PoolConfig

	// client is the Docker API client used to manage containers.
	client *conn.DockerClient

	// idle is a buffered channel containing IDs of idle containers.
	// Borrowing a container reads from this channel; returning a
	// container writes back to it.
	idle chan string

	// mu protects mutations to pool state such as `total`.
	mu sync.Mutex

	// total tracks the total number of containers currently created
	// by the pool (both idle and in-use).
	total int

	// healthMu protects all pool-level health state.
	//
	// It allows concurrent readers (e.g., schedulers, request handlers)
	// while ensuring exclusive access during health updates performed
	// by the background health-check loop.
	healthMu sync.RWMutex

	// poolHealth represents the current aggregated health status of the pool.
	//
	// This value is derived from periodic health checks and reflects
	// the overall readiness of the pool to serve requests.
	health PoolHealth

	// lastHealthCheck records the timestamp of the most recent
	// successful health evaluation.
	//
	// It can be used for observability, debugging, and detecting
	// stalled or delayed health-check loops.
	lastHealthCheck time.Time
}

// NewPool creates and initializes a new ContainerPool for a given language.
//
// It sets up a Docker client using environment configuration and
// prepares an idle container channel sized to the pool's MaxSize.
//
// Note:
//   - This function does not create containers eagerly.
//   - Container creation logic should be handled separately

func NewPool(lang string, cfg PoolConfig) (*ContainerPool, error) {
	cli, err := conn.GetDockerClient()
	if err != nil {
		return nil, err
	}

	return &ContainerPool{
		lang:   lang,
		cfg:    cfg,
		client: cli,
		idle:   make(chan string, cfg.MaxSize),
	}, nil
}
