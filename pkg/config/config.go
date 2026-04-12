package config

import (
	"context"
	"sandman/pkg/grpcutil"
	"time"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/configutil"
)

type Config struct {
	apputil.Config `yaml:",inline"`
	Server         grpcutil.Config `yaml:"server"`
	Worker         WorkerConfig    `yaml:"worker"`
}

// Resolve resolves the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Config.DB.Database, configutil.Lazy(&c.Config.DB.Database), configutil.Const("sandman")),
		(&c.Config).Resolve,
		(&c.Server).Resolve,
	)
}

// WorkerConfig holds configuration shared between workers and the controller.
type WorkerConfig struct {
	BatchSize       int           `yaml:"batch_size"`
	PollingInterval time.Duration `yaml:"polling_interval"`
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
