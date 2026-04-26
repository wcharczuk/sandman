package cron

import (
	"context"
)

// Job is an interface types can satisfy to be loaded into the JobManager.
type Job interface {
	Name() string
	Execute(context.Context) error
}

// ScheduleProvider is a type that provides a schedule for the job.
// If a job does not implement this method, it is treated as
// "OnDemand" or a job that must be triggered explicitly.
type ScheduleProvider interface {
	Schedule() Schedule
}

// ConfigProvider is a type that returns a job config.
type ConfigProvider interface {
	Config() JobConfig
}

// LifecycleProvider is a job that provides lifecycle hooks.
type LifecycleProvider interface {
	Lifecycle() JobLifecycle
}

// BackgroundProvider is a type that returns a base context based on a parent context.
type BackgroundProvider interface {
	Background(context.Context) context.Context
}
