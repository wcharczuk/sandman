package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.charczuk.com/sdk/async"
)

// NewJobScheduler returns a job scheduler for a given job.
func NewJobScheduler(jm *JobManager, job Job) *JobScheduler {
	js := &JobScheduler{
		Job:   job,
		latch: async.NewLatch(),
		jm:    jm,
	}
	if typed, ok := job.(ScheduleProvider); ok {
		js.schedule = typed.Schedule()
	}
	return js
}

// JobScheduler is a job instance.
type JobScheduler struct {
	Job Job

	jm          *JobManager
	config      JobConfig
	schedule    Schedule
	lifecycle   JobLifecycle
	nextRuntime time.Time

	latch       *async.Latch
	currentLock sync.Mutex
	current     *JobInvocation
	lastLock    sync.Mutex
	last        *JobInvocation
}

// Name returns the job name.
func (js *JobScheduler) Name() string {
	return js.Job.Name()
}

// Config returns the job config provided by a job or an empty config.
func (js *JobScheduler) Config() JobConfig {
	if typed, ok := js.Job.(ConfigProvider); ok {
		return typed.Config()
	}
	return js.config
}

// Lifecycle returns job lifecycle steps or an empty set.
func (js *JobScheduler) Lifecycle() JobLifecycle {
	if typed, ok := js.Job.(LifecycleProvider); ok {
		return typed.Lifecycle()
	}
	return js.lifecycle
}

// Labels returns the job labels, including
// automatically added ones like `name`.
func (js *JobScheduler) Labels() map[string]string {
	output := map[string]string{
		"name":      js.Name(),
		"scheduler": string(js.State()),
		"active":    fmt.Sprint(!js.IsIdle()),
		"enabled":   fmt.Sprint(!js.config.Disabled),
	}
	if js.Last() != nil {
		output["last"] = string(js.Last().Status)
	}
	for key, value := range js.Config().Labels {
		output[key] = value
	}
	return output
}

// State returns the job scheduler state.
func (js *JobScheduler) State() JobSchedulerState {
	if js.latch.IsStarted() {
		return JobSchedulerStateRunning
	}
	if js.latch.IsStopped() {
		return JobSchedulerStateStopped
	}
	return JobSchedulerStateUnknown
}

// Start starts the scheduler.
// This call blocks.
func (js *JobScheduler) Start(ctx context.Context) error {
	if !js.latch.CanStart() {
		return ErrCannotStart
	}
	js.latch.Starting()
	js.runLoop(ctx)
	return nil
}

// Stop stops the scheduler.
func (js *JobScheduler) Stop(ctx context.Context) error {
	if !js.latch.CanStop() {
		return ErrCannotStop
	}
	js.latch.Stopping() // trigger the `NotifyStopping` channel

	// if it's currently running
	// cancel or wait to cancel
	if current := js.Current(); current != nil {
		ctx := js.withBaseContext(ctx)
		gracePeriod := js.Config().ShutdownGracePeriodOrDefault()
		if gracePeriod > 0 {
			var cancel func()
			ctx, cancel = js.withTimeoutOrCancel(ctx, gracePeriod)
			defer cancel()
			js.waitCurrentComplete(ctx)
		} else {
			current.Cancel()
		}
	}

	// wait for the runloop to exit
	<-js.latch.NotifyStopped()
	js.latch.Reset()
	js.nextRuntime = Zero
	return nil
}

// NotifyStarted notifies the job scheduler has started.
func (js *JobScheduler) NotifyStarted() <-chan struct{} {
	return js.latch.NotifyStarted()
}

// NotifyStopped notifies the job scheduler has stopped.
func (js *JobScheduler) NotifyStopped() <-chan struct{} {
	return js.latch.NotifyStopped()
}

// Enable sets the job as enabled.
func (js *JobScheduler) Enable(ctx context.Context) {
	ctx = js.withBaseContext(ctx)
	js.config.Disabled = false
	if lifecycle := js.Lifecycle(); lifecycle.OnEnabled != nil {
		lifecycle.OnEnabled(ctx)
	}
}

// Disable sets the job as disabled.
func (js *JobScheduler) Disable(ctx context.Context) {
	ctx = js.withBaseContext(ctx)
	js.config.Disabled = true
	if lifecycle := js.Lifecycle(); lifecycle.OnDisabled != nil {
		lifecycle.OnDisabled(ctx)
	}
}

// Cancel stops all running invocations.
func (js *JobScheduler) Cancel(ctx context.Context) error {
	ctx = js.withBaseContext(ctx)
	if js.Current() == nil {
		return nil
	}
	gracePeriod := js.Config().ShutdownGracePeriodOrDefault()
	if gracePeriod > 0 {
		ctx, cancel := js.withTimeoutOrCancel(ctx, gracePeriod)
		defer cancel()
		js.waitCurrentComplete(ctx)
	}
	if current := js.Current(); current != nil && current.Status == JobInvocationStatusRunning {
		current.Cancel()
	}
	return nil
}

// RunAsync starts a job invocation with a given context.
func (js *JobScheduler) RunAsync(ctx context.Context) (*JobInvocation, error) {
	if !js.IsIdle() {
		return nil, ErrJobAlreadyRunning
	}

	ctx = js.withBaseContext(ctx)
	ctx, ji := js.withInvocationContext(ctx)
	js.setCurrent(ji)

	var err error
	go func() {
		defer func() {
			switch {
			case err != nil && IsJobCanceled(err):
				js.onJobCompleteCanceled(ctx) // the job was canceled, either manually or by a timeout
			case err != nil:
				js.onJobCompleteError(ctx, err) // the job completed with an error
			default:
				js.onJobCompleteSuccess(ctx) // the job completed without error
			}
			ji.Cancel()              // if the job was created with a timeout, end the timeout
			js.assignCurrentToLast() // rotate in the current to the last result
		}()
		js.onJobBegin(ctx) // signal the job is starting

		select {
		case <-ctx.Done(): // if the timeout or cancel is triggered
			err = ErrJobCanceled // set the error to a known error
			return
		case err = <-js.safeBackgroundExec(ctx): // run the job in a background routine and catch panics
			return
		}
	}()
	return ji, nil
}

// Run forces the job to run.
// This call will block.
func (js *JobScheduler) Run(ctx context.Context) {
	ji, err := js.RunAsync(ctx)
	if err != nil {
		return
	}
	<-ji.Finished
}

//
// exported utility methods
//

// CanBeScheduled returns if a job will be triggered automatically
// and isn't already in flight and set to be serial.
func (js *JobScheduler) CanBeScheduled() bool {
	return !js.config.Disabled && js.IsIdle()
}

// IsIdle returns if the job is not currently running.
func (js *JobScheduler) IsIdle() (isIdle bool) {
	isIdle = js.Current() == nil
	return
}

//
// internal functions
//

func (js *JobScheduler) runLoop(ctx context.Context) {
	js.latch.Started()
	defer func() {
		js.latch.Stopped()
		js.latch.Reset()
	}()

	if js.schedule != nil {
		js.nextRuntime = js.schedule.Next(js.nextRuntime)
	}
	if js.nextRuntime.IsZero() {
		return
	}

	runAt := time.NewTimer(js.nextRuntime.UTC().Sub(Now()))
	for {
		select {
		case <-runAt.C:
			runAt.Stop()
			if js.CanBeScheduled() {
				_, _ = js.RunAsync(ctx)
			}
			if !js.latch.IsStarted() {
				return
			}
			if js.schedule != nil {
				js.nextRuntime = js.schedule.Next(js.nextRuntime)
				runAt.Reset(js.nextRuntime.UTC().Sub(Now()))
			} else {
				js.nextRuntime = Zero
			}
			if js.nextRuntime.IsZero() {
				return
			}
		case <-js.latch.NotifyStopping():
			runAt.Stop()
			return
		}
	}
}

// Current returns the current job invocation.
func (js *JobScheduler) Current() (current *JobInvocation) {
	js.currentLock.Lock()
	if js.current != nil {
		current = js.current.Clone()
	}
	js.currentLock.Unlock()
	return
}

// Last returns the last job invocation.
func (js *JobScheduler) Last() (last *JobInvocation) {
	js.lastLock.Lock()
	if js.last != nil {
		last = js.last
	}
	js.lastLock.Unlock()
	return
}

// SetCurrent sets the current invocation, it is useful for tests etc.
func (js *JobScheduler) setCurrent(ji *JobInvocation) {
	js.currentLock.Lock()
	js.current = ji
	js.currentLock.Unlock()
}

// SetLast sets the last invocation, it is useful for tests etc.
func (js *JobScheduler) setLast(ji *JobInvocation) {
	js.lastLock.Lock()
	js.last = ji
	js.lastLock.Unlock()
}

func (js *JobScheduler) onRegister(ctx context.Context) error {
	ctx = js.withBaseContext(ctx)
	if js.Lifecycle().OnRegister != nil {
		if err := js.Lifecycle().OnRegister(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (js *JobScheduler) onRemove(ctx context.Context) error {
	ctx = js.withBaseContext(ctx)
	if js.Lifecycle().OnRemove != nil {
		return js.Lifecycle().OnRemove(ctx)
	}
	return nil
}

func (js *JobScheduler) assignCurrentToLast() {
	js.lastLock.Lock()
	js.currentLock.Lock()
	js.last = js.current
	js.current = nil
	js.currentLock.Unlock()
	js.lastLock.Unlock()
}

func (js *JobScheduler) waitCurrentComplete(ctx context.Context) {
	if js.Current().Status != JobInvocationStatusRunning {
		return
	}

	finished := js.current.Finished
	select {
	case <-ctx.Done():
		js.Current().Cancel()
		return
	case <-finished:
		// tick over the loop to check if the current job is complete
		return
	}
}

func (js *JobScheduler) safeBackgroundExec(ctx context.Context) <-chan error {
	errors := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errors <- fmt.Errorf("%v", r)
			}
		}()
		errors <- js.Job.Execute(ctx)
	}()
	return errors
}

func (js *JobScheduler) withBaseContext(ctx context.Context) context.Context {
	if typed, ok := js.Job.(BackgroundProvider); ok {
		ctx = typed.Background(ctx)
	}
	ctx = WithJobScheduler(ctx, js)
	return ctx
}

func (js *JobScheduler) withTimeoutOrCancel(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return context.WithCancel(ctx)
}

func (js *JobScheduler) withInvocationContext(ctx context.Context) (context.Context, *JobInvocation) {
	ji := newJobInvocation(js.Name())
	ji.Parameters = MergeJobParameterValues(js.Config().ParameterValues, GetJobParameterValues(ctx))
	ctx, ji.Cancel = js.withTimeoutOrCancel(ctx, js.Config().TimeoutOrDefault())
	ctx = WithJobInvocation(ctx, ji)
	ctx = WithJobParameterValues(ctx, ji.Parameters)
	return ctx, ji
}

// job lifecycle hooks

func (js *JobScheduler) onJobBegin(ctx context.Context) {
	js.currentLock.Lock()
	js.current.Started = time.Now().UTC()
	js.current.Status = JobInvocationStatusRunning
	js.currentLock.Unlock()

	if lifecycle := js.Lifecycle(); lifecycle.OnBegin != nil {
		lifecycle.OnBegin(ctx)
	}
	if js.jm != nil && len(js.jm.onJobBegin) > 0 {
		jse := JobSchedulerEvent{
			Phase:         "job.begin",
			JobName:       js.Job.Name(),
			JobInvocation: GetJobInvocation(ctx).ID,
			Parameters:    GetJobParameterValues(ctx),
		}
		for _, listener := range js.jm.onJobBegin {
			listener(jse)
		}
	}
}

func (js *JobScheduler) onJobCompleteCanceled(ctx context.Context) {
	js.currentLock.Lock()
	js.current.Complete = time.Now().UTC()
	js.current.Status = JobInvocationStatusCanceled
	close(js.current.Finished)
	js.currentLock.Unlock()

	lifecycle := js.Lifecycle()
	if lifecycle.OnCancellation != nil {
		lifecycle.OnCancellation(ctx)
	}
	if lifecycle.OnComplete != nil {
		lifecycle.OnComplete(ctx)
	}
	if js.jm != nil && len(js.jm.onJobComplete) > 0 {
		jse := JobSchedulerEvent{
			Phase:         "job.canceled",
			JobName:       js.Job.Name(),
			JobInvocation: GetJobInvocation(ctx).ID,
			Parameters:    GetJobParameterValues(ctx),
			Elapsed:       js.current.Complete.Sub(js.current.Started),
		}
		for _, listener := range js.jm.onJobComplete {
			listener(jse)
		}
	}
}

func (js *JobScheduler) onJobCompleteSuccess(ctx context.Context) {
	js.currentLock.Lock()
	js.current.Complete = time.Now().UTC()
	js.current.Status = JobInvocationStatusSuccess
	close(js.current.Finished)
	js.currentLock.Unlock()

	lifecycle := js.Lifecycle()
	if lifecycle.OnSuccess != nil {
		lifecycle.OnSuccess(ctx)
	}
	if last := js.Last(); last != nil && last.Status == JobInvocationStatusErrored {
		if lifecycle.OnFixed != nil {
			lifecycle.OnFixed(ctx)
		}
	}
	if lifecycle.OnComplete != nil {
		lifecycle.OnComplete(ctx)
	}
	if js.jm != nil && len(js.jm.onJobComplete) > 0 {
		jse := JobSchedulerEvent{
			Phase:         "job.complete",
			JobName:       js.Job.Name(),
			JobInvocation: GetJobInvocation(ctx).ID,
			Parameters:    GetJobParameterValues(ctx),
			Elapsed:       js.current.Complete.Sub(js.current.Started),
		}
		for _, listener := range js.jm.onJobComplete {
			listener(jse)
		}
	}
}

func (js *JobScheduler) onJobCompleteError(ctx context.Context, err error) {
	js.currentLock.Lock()
	js.current.Complete = time.Now().UTC()
	js.current.Status = JobInvocationStatusErrored
	js.current.Err = err
	close(js.current.Finished)
	js.currentLock.Unlock()

	//
	// error
	//

	// always log the error
	lifecycle := js.Lifecycle()
	if lifecycle.OnError != nil {
		lifecycle.OnError(ctx)
	}

	//
	// broken; assumes that last is set, and last was a success
	//

	if last := js.Last(); last != nil && last.Status != JobInvocationStatusErrored {
		if lifecycle.OnBroken != nil {
			lifecycle.OnBroken(ctx)
		}
	}
	if lifecycle.OnComplete != nil {
		lifecycle.OnComplete(ctx)
	}
	if js.jm != nil && len(js.jm.onJobComplete) > 0 {
		jse := JobSchedulerEvent{
			Phase:         "job.error",
			JobName:       js.Job.Name(),
			JobInvocation: GetJobInvocation(ctx).ID,
			Parameters:    GetJobParameterValues(ctx),
			Elapsed:       js.current.Complete.Sub(js.current.Started),
			Err:           err,
		}
		for _, listener := range js.jm.onJobComplete {
			listener(jse)
		}
	}
}
