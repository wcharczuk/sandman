package cron

import (
	"context"
	"sync"

	"go.charczuk.com/sdk/async"
	"go.charczuk.com/sdk/log"
)

// New returns a new job manager.
func New(opts ...Option) *JobManager {
	jm := JobManager{
		latch: async.NewLatch(),
		jobs:  make(map[string]*JobScheduler),
	}
	for _, opt := range opts {
		opt(&jm)
	}
	return &jm
}

// Option mutates a job manager.
type Option func(*JobManager)

// OptLog registers lifecycle hooks to emit log messages.
func OptLog(log *log.Logger) Option {
	return func(jm *JobManager) {
		jm.onStart = append(jm.onStart, func() {
			log.WithGroup("CRON").Info("job manager starting")
		})
		jm.onStop = append(jm.onStop, func() {
			log.WithGroup("CRON").Info("job manager stopping")
		})
		jm.onJobBegin = append(jm.onJobBegin, func(e JobSchedulerEvent) {
			log.WithGroup("CRON").Info("job_starting", e)
		})
		jm.onJobComplete = append(jm.onJobComplete, func(e JobSchedulerEvent) {
			log.WithGroup("CRON").Info("job_complete", e)
			if e.Err != nil {
				log.WithGroup("cron").Error("job_err", e.Err)
			}
		})
	}
}

// JobManager is the main orchestration and job management object.
type JobManager struct {
	mu    sync.Mutex
	latch *async.Latch
	jobs  map[string]*JobScheduler

	onStart []func()
	onStop  []func()

	onJobBegin    []func(JobSchedulerEvent)
	onJobComplete []func(JobSchedulerEvent)
}

//
// Life Cycle
//

// Start starts the job manager and blocks.
func (jm *JobManager) Start(ctx context.Context) error {
	if err := jm.StartAsync(ctx); err != nil {
		return err
	}
	<-jm.latch.NotifyStopped()
	return nil
}

// StartAsync starts the job manager and the loaded jobs.
// It does not block.
func (jm *JobManager) StartAsync(ctx context.Context) error {
	if !jm.latch.CanStart() {
		return ErrCannotStart
	}
	jm.latch.Starting()
	for _, jobScheduler := range jm.jobs {
		go jobScheduler.Start(ctx)
		<-jobScheduler.NotifyStarted()
	}
	for _, hook := range jm.onStart {
		hook()
	}
	jm.latch.Started()
	return nil
}

// Restart doesn't do anything right now.
func (jm *JobManager) Restart(ctx context.Context) error {
	return nil
}

// Stop stops the schedule runner for a JobManager.
func (jm *JobManager) Stop(ctx context.Context) error {
	if !jm.latch.CanStop() {
		return ErrCannotStop
	}
	for _, hook := range jm.onStop {
		hook()
	}
	jm.latch.Stopping()
	defer func() {
		jm.latch.Stopped()
		jm.latch.Reset()
	}()
	for _, jobScheduler := range jm.jobs {
		_ = jobScheduler.onRemove(ctx)
		_ = jobScheduler.Stop(ctx)
	}
	return nil
}

//
// job management
//

// Register adds list of jobs to the job manager and calls their
// "OnRegister" lifecycle handler(s).
func (jm *JobManager) Register(ctx context.Context, jobs ...Job) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	for _, job := range jobs {
		jobName := job.Name()
		if _, hasJob := jm.jobs[jobName]; hasJob {
			return ErrJobAlreadyLoaded
		}
		jobScheduler := NewJobScheduler(jm, job)
		if err := jobScheduler.onRegister(ctx); err != nil {
			return err
		}
		jm.jobs[jobName] = jobScheduler
	}
	return nil
}

// Remove removes jobs from the manager and stops them.
func (jm *JobManager) Remove(ctx context.Context, jobNames ...string) (err error) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	for _, jobName := range jobNames {
		if jobScheduler, ok := jm.jobs[jobName]; ok {
			err = jobScheduler.onRemove(ctx)
			if err != nil {
				return
			}
			err = jobScheduler.Stop(ctx)
			if err != nil && err != ErrCannotStop {
				return
			}
			delete(jm.jobs, jobName)
		} else {
			return ErrJobNotLoaded
		}
	}
	return nil
}

// Disable disables a variadic list of job names.
func (jm *JobManager) Disable(ctx context.Context, jobNames ...string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	for _, jobName := range jobNames {
		if job, ok := jm.jobs[jobName]; ok {
			job.Disable(ctx)
		} else {
			return ErrJobNotLoaded
		}
	}
	return nil
}

// Enable enables a variadic list of job names.
func (jm *JobManager) Enable(ctx context.Context, jobNames ...string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	for _, jobName := range jobNames {
		if job, ok := jm.jobs[jobName]; ok {
			job.Enable(ctx)
		} else {
			return ErrJobNotLoaded
		}
	}
	return nil
}

// Has returns if a jobName is loaded or not.
func (jm *JobManager) Has(jobName string) (hasJob bool) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	_, hasJob = jm.jobs[jobName]
	return
}

// Job returns a job metadata by name.
func (jm *JobManager) Job(jobName string) (job *JobScheduler, ok bool) {
	jm.mu.Lock()
	job, ok = jm.jobs[jobName]
	jm.mu.Unlock()
	return
}

// IsJobDisabled returns if a job is disabled.
func (jm *JobManager) IsJobDisabled(jobName string) (value bool) {
	jm.mu.Lock()
	jobScheduler, hasJob := jm.jobs[jobName]
	jm.mu.Unlock()
	if hasJob {
		value = jobScheduler.Config().Disabled
	}
	return
}

// IsJobRunning returns if a job is currently running.
func (jm *JobManager) IsJobRunning(jobName string) (isRunning bool) {
	jm.mu.Lock()
	jobScheduler, ok := jm.jobs[jobName]
	jm.mu.Unlock()
	if ok {
		isRunning = !jobScheduler.IsIdle()
	}
	return
}

// RunJob runs a job by jobName on demand with a given context.
func (jm *JobManager) RunJob(ctx context.Context, jobName string) (*JobInvocation, error) {
	jm.mu.Lock()
	jobScheduler, ok := jm.jobs[jobName]
	jm.mu.Unlock()
	if !ok {
		return nil, ErrJobNotLoaded
	}
	return jobScheduler.RunAsync(ctx)
}

// CancelJob cancels (sends the cancellation signal) to a running job.
func (jm *JobManager) CancelJob(ctx context.Context, jobName string) (err error) {
	jm.mu.Lock()
	jobScheduler, ok := jm.jobs[jobName]
	jm.mu.Unlock()
	if !ok {
		err = ErrJobNotLoaded
		return
	}
	err = jobScheduler.Cancel(ctx)
	return
}

//
// status and state
//

// State returns the job manager state.
func (jm *JobManager) State() JobManagerState {
	if jm.latch.IsStarted() {
		return JobManagerStateRunning
	} else if jm.latch.IsStopped() {
		return JobManagerStateStopped
	}
	return JobManagerStateUnknown
}
