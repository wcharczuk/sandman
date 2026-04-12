package control

import "time"

// Config holds the configuration for the scale controller.
type Config struct {
	EvaluationInterval    time.Duration
	MinReplicas           int32
	MaxReplicas           int32
	WorkerBatchSize       int
	WorkerPollingInterval time.Duration
}

const (
	DefaultEvaluationInterval          = 5 * time.Minute
	DefaultMinReplicas       int32     = 3
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

// K8sConfig holds Kubernetes-specific controller settings.
type K8sConfig struct {
	Namespace  string `yaml:"namespace"`
	Deployment string `yaml:"deployment"`
	LeaseName  string `yaml:"lease_name"`
	PodName    string `yaml:"pod_name"`
}
