package control

import "time"

// Config holds the configuration for the scale controller.
type Config struct {
	EvaluationInterval    time.Duration
	MinReplicas           int32
	MaxReplicas           int32
	WorkerBatchSize       int
	WorkerPollingInterval time.Duration
	// CullInterval is how often delivered/exhausted timers are swept
	// from the table. Independent of the scaling loop so a slow cull
	// never delays scale decisions.
	CullInterval time.Duration
	// CullRetention is how long a delivered timer is kept past its
	// due_utc before it becomes eligible for deletion. Gives operators
	// a grace period to inspect recently-delivered rows before they go.
	CullRetention time.Duration
}

const (
	DefaultEvaluationInterval          = 5 * time.Minute
	DefaultMinReplicas       int32     = 3
	DefaultCullInterval                = 1 * time.Minute
	DefaultCullRetention               = 5 * time.Minute
)

func (c Config) EvaluationIntervalOrDefault() time.Duration {
	if c.EvaluationInterval > 0 {
		return c.EvaluationInterval
	}
	return DefaultEvaluationInterval
}

func (c Config) MinReplicasOrDefault() int32 {
	if c.MinReplicas > 0 {
		return c.MinReplicas
	}
	return DefaultMinReplicas
}

func (c Config) WorkerBatchSizeOrDefault() int {
	if c.WorkerBatchSize > 0 {
		return c.WorkerBatchSize
	}
	return 1024
}

func (c Config) WorkerPollingIntervalOrDefault() time.Duration {
	if c.WorkerPollingInterval > 0 {
		return c.WorkerPollingInterval
	}
	return 5 * time.Second
}

func (c Config) CullIntervalOrDefault() time.Duration {
	if c.CullInterval > 0 {
		return c.CullInterval
	}
	return DefaultCullInterval
}

func (c Config) CullRetentionOrDefault() time.Duration {
	if c.CullRetention > 0 {
		return c.CullRetention
	}
	return DefaultCullRetention
}

// K8sConfig holds Kubernetes-specific controller settings.
type K8sConfig struct {
	Namespace  string `yaml:"namespace"`
	Deployment string `yaml:"deployment"`
	LeaseName  string `yaml:"lease_name"`
	PodName    string `yaml:"pod_name"`
}
