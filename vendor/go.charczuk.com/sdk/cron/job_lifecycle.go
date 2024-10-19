package cron

import "context"

// JobLifecycle is a suite of lifeycle hooks
// you can set for a given job.
type JobLifecycle struct {
	// OnRegister is called when the job is loaded into the job manager.
	OnRegister func(context.Context) error
	// OnRemove is called when the job is removed from the manager
	// or the job manager is stopped.
	OnRemove func(context.Context) error
	// OnBegin fires whenever a job is started.
	OnBegin func(context.Context)
	// OnComplete fires whenever a job finishes, regardless of status.
	OnComplete func(context.Context)
	// OnCancellation is called if the job is canceled explicitly
	// or it sets a timeout in the .Config() and exceeds that timeout.
	OnCancellation func(context.Context)
	// OnError is called if the job returns an error or panics during
	// execution, but will not be called if the job is canceled.
	OnError func(context.Context)
	// OnSuccess is called if the job completes without an error.
	OnSuccess func(context.Context)
	// OnBroken is called if the job errors after having completed successfully
	// the previous invocation.
	OnBroken func(context.Context)
	// OnFixed is called if the job completes successfully after having
	// returned an error on the previous invocation.
	OnFixed func(context.Context)
	// OnEnabled is called if the job is explicitly enabled.
	OnEnabled func(context.Context)
	// OnDisabled is called if the job is explicitly disabled.
	OnDisabled func(context.Context)
}
