package config

import (
	"context"
	"sandman/pkg/grpcutil"
	"strings"
	"time"

	"sandman/pkg/apputil"
	"sandman/pkg/configutil"
)

type Config struct {
	apputil.Config `yaml:",inline"`
	// DBHosts is a list of `host:port` entries. When non-empty, the list
	// overrides db.host / db.port with a comma-joined multi-host string and
	// turns on db.loadBalanceHosts so the pool shuffles which node each new
	// connection lands on.
	DBHosts []string        `yaml:"dbHosts,omitempty"`
	Server  grpcutil.Config `yaml:"server"`
	Worker  WorkerConfig    `yaml:"worker"`
}

// DefaultDBMaxLifetime bounds how long a pooled connection sticks to one
// node when host shuffling is enabled. Without this, connections that
// happened to land on n1 at pool warm-up stay there for the process
// lifetime, defeating the rebalance hook.
const DefaultDBMaxLifetime = 5 * time.Minute

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	if err := configutil.Resolve(ctx,
		configutil.Set(&c.Config.DB.Database, configutil.Lazy(&c.Config.DB.Database), configutil.Const("sandman")),
		(&c.Config).Resolve,
		(&c.Server).Resolve,
	); err != nil {
		return err
	}
	// Apply the multi-host override AFTER apputil.Resolve: the sdk's
	// db.Config.Resolve defaults DB.Port to "5432" when unset, which would
	// get appended to the multi-host string and produce an invalid DSN.
	if len(c.DBHosts) > 0 {
		c.Config.DB.Host = strings.Join(c.DBHosts, ",")
		c.Config.DB.Port = ""
		c.Config.DB.LoadBalanceHosts = true
		if c.Config.DB.MaxLifetime == 0 {
			c.Config.DB.MaxLifetime = DefaultDBMaxLifetime
		}
	}
	return nil
}

// WorkerConfig holds configuration shared between workers and the controller.
type WorkerConfig struct {
	BatchSize       int           `yaml:"batch_size"`
	PollingInterval time.Duration `yaml:"polling_interval"`
	// PrefetchWindow turns on wheel-mode dispatch. Zero (default)
	// keeps the original "claim due now, fire, mark delivered" tick
	// loop. Anything > 0 swaps in three loops (prefetch, dispatch,
	// flush) keyed off an in-memory hash wheel.
	PrefetchWindow       time.Duration `yaml:"prefetch_window"`
	DispatchTickInterval time.Duration `yaml:"dispatch_tick_interval"`
	FlushInterval        time.Duration `yaml:"flush_interval"`
}

const (
	DefaultWorkerBatchSize                        = 255
	DefaultWorkerPollingInterval time.Duration = 5 * time.Second
)

func (wc WorkerConfig) BatchSizeOrDefault() int {
	if wc.BatchSize > 0 {
		return wc.BatchSize
	}
	return DefaultWorkerBatchSize
}

func (wc WorkerConfig) PollingIntervalOrDefault() time.Duration {
	if wc.PollingInterval > 0 {
		return wc.PollingInterval
	}
	return DefaultWorkerPollingInterval
}
