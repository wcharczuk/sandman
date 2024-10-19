package async

import (
	"context"
	"sync"
	"time"
)

/*
NewInterval returns a new worker that runs an action on an interval.

Example:

	iw := &Interval{
		Action: func(ctx context.Context) error {
			fmt.Println("running!")
			return nil
		},
		Interval: 500*time.Millisecond,
	}
	go iw.Start()
	<-iw.Started()

*/

// Interval is a background worker that performs an action on an interval.
type Interval struct {
	Context     context.Context
	Interval    time.Duration
	Action      func(context.Context) error
	Delay       time.Duration
	StopOnError bool
	Errors      chan error

	latchMu sync.Mutex
	latch   *Latch
}

// Interval defaults
const (
	DefaultInterval = 500 * time.Millisecond
)

// Background returns the provided context or context.Background()
func (i *Interval) Background() context.Context {
	if i.Context != nil {
		return i.Context
	}
	return context.Background()
}

// IntervalOrDefault returns the interval or a default.
func (i *Interval) IntervalOrDefault() time.Duration {
	if i.Interval > 0 {
		return i.Interval
	}
	return DefaultInterval
}

// Started returns the channel to notify when the worker starts.
func (i *Interval) Started() <-chan struct{} {
	return i.latch.NotifyStarted()
}

/*
Start starts the worker.

This will start the internal ticker, with a default initial delay of the given interval, and will return an ErrCannotStart if the interval worker is already started.

This call will block.
*/
func (i *Interval) Start() error {
	i.ensureLatch()
	if !i.latch.CanStart() {
		return ErrCannotStart
	}
	if i.Action == nil {
		return ErrCannotStartActionRequired
	}
	i.latch.Starting()
	return i.Dispatch()
}

// Stop stops the worker.
func (i *Interval) Stop() error {
	if !i.latch.CanStop() {
		return ErrCannotStop
	}
	i.latch.Stopping()
	<-i.latch.NotifyStopped()
	i.latch.Reset() // reset the latch in case we have to start again
	return nil
}

// Dispatch is the main dispatch loop.
func (i *Interval) Dispatch() (err error) {
	i.latch.Started()

	if i.Delay > 0 {
		alarm := time.NewTimer(i.Delay)
		stopping := i.latch.NotifyStopping()
		select {
		case <-i.Context.Done():
			alarm.Stop()
			return
		case <-stopping:
			alarm.Stop()
			return
		case <-alarm.C:
			alarm.Stop()
		}
	}

	tick := time.NewTicker(i.IntervalOrDefault())
	defer func() {
		tick.Stop()
		i.latch.Stopped()
	}()

	var stopping <-chan struct{}
	for {
		stopping = i.latch.NotifyStopping()
		select {
		case <-i.Context.Done():
			return
		case <-stopping:
			return
		default:
		}

		select {
		case <-i.Context.Done():
			return
		case <-stopping:
			return
		case <-tick.C:
			err = i.Action(i.Background())
			if err != nil {
				if i.StopOnError {
					return
				}
				if i.Errors != nil {
					select {
					case <-stopping:
						return
					case <-i.Context.Done():
						return
					case i.Errors <- err:
					}
				}
			}
		}
	}
}

func (i *Interval) ensureLatch() {
	i.latchMu.Lock()
	defer i.latchMu.Unlock()
	if i.latch == nil {
		i.latch = NewLatch()
	}
}
