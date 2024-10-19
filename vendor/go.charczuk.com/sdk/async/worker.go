package async

import (
	"context"
	"fmt"
)

// NewWorker creates a new worker.
func NewWorker[T any](action WorkAction[T]) *Worker[T] {
	return &Worker[T]{
		Latch:   NewLatch(),
		Context: context.Background(),
		Action:  action,
		Work:    make(chan T),
	}
}

// WorkAction is an action handler for a queue.
type WorkAction[T any] func(context.Context, T) error

// WorkerFinalizer is an action handler for a queue.
type WorkerFinalizer[T any] func(context.Context, *Worker[T]) error

// Worker is a worker that is pushed work over a channel.
// It is used by other work distribution types (i.e. queue and batch)
// but can also be used independently.
type Worker[T any] struct {
	*Latch

	Context   context.Context
	Action    WorkAction[T]
	Finalizer WorkerFinalizer[T]

	SkipRecover bool
	Errors      chan error
	Work        chan T
}

// Background returns the queue worker background context.
func (w *Worker[T]) Background() context.Context {
	if w.Context != nil {
		return w.Context
	}
	return context.Background()
}

// NotifyStarted returns the underlying latch signal.
func (w *Worker[T]) NotifyStarted() <-chan struct{} {
	return w.Latch.NotifyStarted()
}

// NotifyStopped returns the underlying latch signal.
func (w *Worker[T]) NotifyStopped() <-chan struct{} {
	return w.Latch.NotifyStarted()
}

// Start starts the worker with a given context.
func (w *Worker[T]) Start() error {
	if !w.Latch.CanStart() {
		return ErrCannotStart
	}
	w.Latch.Starting()
	w.Dispatch()
	return nil
}

// Dispatch starts the listen loop for work.
func (w *Worker[T]) Dispatch() {
	w.Latch.Started()
	defer w.Latch.Stopped()

	var workItem T
	var stopping <-chan struct{}
	for {
		stopping = w.Latch.NotifyStopping()
		select {
		case <-stopping:
			return
		case <-w.Background().Done():
			return
		default:
		}

		// block on work or stopping
		select {
		case workItem = <-w.Work:
			w.Execute(w.Background(), workItem)
		case <-stopping:
			return
		case <-w.Background().Done():
			return
		}
	}
}

// Execute invokes the action and recovers panics.
func (w *Worker[T]) Execute(ctx context.Context, workItem T) {
	defer func() {
		if !w.SkipRecover {
			if r := recover(); r != nil {
				w.HandlePanic(r)
			}
		}
		if w.Finalizer != nil {
			w.HandleError(w.Finalizer(ctx, w))
		}
	}()
	if w.Action != nil {
		w.HandleError(w.Action(ctx, workItem))
	}
}

// Stop stops the worker.
//
// If there is an item left in the work channel
// it will be processed synchronously.
func (w *Worker[T]) Stop() error {
	if !w.Latch.CanStop() {
		return ErrCannotStop
	}
	w.Latch.WaitStopped()
	w.Latch.Reset()
	return nil
}

// StopContext stops the worker in a given cancellation context.
func (w *Worker[T]) StopContext(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		defer func() {
			w.Latch.Reset()
			close(stopped)
		}()

		w.Latch.WaitStopped()
		if workLeft := len(w.Work); workLeft > 0 {
			for x := 0; x < workLeft; x++ {
				w.Execute(ctx, <-w.Work)
			}
		}
	}()
	select {
	case <-stopped:
		return
	case <-ctx.Done():
		return
	}
}

// HandleError sends a non-nil err to the error
// collector if one is provided.
func (w *Worker[T]) HandlePanic(r interface{}) {
	if r == nil {
		return
	}
	if w.Errors == nil {
		return
	}
	w.Errors <- fmt.Errorf("%v", r)
}

// HandleError sends a non-nil err to the error
// collector if one is provided.
func (w *Worker[T]) HandleError(err error) {
	if err == nil {
		return
	}
	if w.Errors == nil {
		return
	}
	w.Errors <- err
}
