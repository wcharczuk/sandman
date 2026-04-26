package cron

import (
	"time"
)

// Constats and defaults
const (
	DefaultTimeout             time.Duration = 0
	DefaultShutdownGracePeriod time.Duration = 0
)

// JobManagerState is a job manager status.
type JobManagerState string

// JobManagerState values.
const (
	JobManagerStateUnknown JobManagerState = "unknown"
	JobManagerStateRunning JobManagerState = "started"
	JobManagerStateStopped JobManagerState = "stopped"
)

// JobSchedulerState is a job manager status.
type JobSchedulerState string

// JobManagerState values.
const (
	JobSchedulerStateUnknown JobSchedulerState = "unknown"
	JobSchedulerStateRunning JobSchedulerState = "started"
	JobSchedulerStateStopped JobSchedulerState = "stopped"
)

// JobInvocationStatus is a job status.
type JobInvocationStatus string

// JobInvocationState values.
const (
	JobInvocationStatusIdle     JobInvocationStatus = "idle"
	JobInvocationStatusRunning  JobInvocationStatus = "running"
	JobInvocationStatusCanceled JobInvocationStatus = "canceled"
	JobInvocationStatusErrored  JobInvocationStatus = "errored"
	JobInvocationStatusSuccess  JobInvocationStatus = "success"
)
