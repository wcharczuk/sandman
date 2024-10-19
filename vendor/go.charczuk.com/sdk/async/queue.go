package async

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// Queue is a queue with multiple workers.
type Queue[T any] struct {
	Action              WorkAction[T]
	Context             context.Context
	Errors              chan error
	Parallelism         int
	MaxWork             int
	ShutdownGracePeriod time.Duration

	Work chan T

	latchMu          sync.Mutex
	latch            *Latch
	availableWorkers chan *Worker[T]
	workers          []*Worker[T]
}

// Queue Constants
const (
	DefaultQueueMaxWork             = 1 << 10
	DefaultQueueShutdownGracePeriod = 10 * time.Second
)

// Background returns a background context.
func (q *Queue[T]) Background() context.Context {
	if q.Context != nil {
		return q.Context
	}
	return context.Background()
}

// ParallelismOrDefault returns the queue worker parallelism
// or a default value, which is the number of CPUs.
func (q *Queue[T]) ParallelismOrDefault() int {
	if q.Parallelism > 0 {
		return q.Parallelism
	}
	return runtime.NumCPU()
}

// MaxWorkOrDefault returns the work queue capacity
// or a default value if it is unset.
func (q *Queue[T]) MaxWorkOrDefault() int {
	if q.MaxWork > 0 {
		return q.MaxWork
	}
	return DefaultQueueMaxWork
}

// ShutdownGracePeriodOrDefault returns the work queue shutdown grace period
// or a default value if it is unset.
func (q *Queue[T]) ShutdownGracePeriodOrDefault() time.Duration {
	if q.ShutdownGracePeriod > 0 {
		return q.ShutdownGracePeriod
	}
	return DefaultQueueShutdownGracePeriod
}

// Start starts the queue and its workers.
// This call blocks.
func (q *Queue[T]) Start() error {
	q.ensureLatch()

	if !q.latch.CanStart() {
		return ErrCannotStart
	}
	if q.Action == nil {
		return ErrCannotStartActionRequired
	}
	q.latch.Starting()

	q.Work = make(chan T, q.MaxWorkOrDefault())
	q.availableWorkers = make(chan *Worker[T], q.ParallelismOrDefault())
	q.workers = make([]*Worker[T], q.ParallelismOrDefault())

	for x := 0; x < q.ParallelismOrDefault(); x++ {
		worker := NewWorker(q.Action)
		worker.Context = q.Context
		worker.Errors = q.Errors
		worker.Finalizer = q.returnWorker

		// start the worker on its own goroutine
		go func() { _ = worker.Start() }()
		<-worker.NotifyStarted()
		q.availableWorkers <- worker
		q.workers[x] = worker
	}
	q.Dispatch()
	return nil
}

// Dispatch processes work items in a loop.
func (q *Queue[T]) Dispatch() {
	q.latch.Started()
	defer q.latch.Stopped()

	var workItem T
	var worker *Worker[T]
	var stopping <-chan struct{}
	for {
		stopping = q.latch.NotifyStopping()
		select {
		case <-stopping:
			return
		case <-q.Background().Done():
			return
		default:
		}

		select {
		case <-stopping:
			return
		case <-q.Background().Done():
			return
		case workItem = <-q.Work:
			select {
			case <-stopping:
				q.Work <- workItem
				return
			case <-q.Background().Done():
				q.Work <- workItem
				return
			case worker = <-q.availableWorkers:
				select {
				case <-stopping:
					q.Work <- workItem
					return
				case <-q.Background().Done():
					q.Work <- workItem
					return
				case worker.Work <- workItem:
				}
			}
		}
	}
}

// Stop stops the queue and processes any remaining items.
func (q *Queue[T]) Stop() error {
	if !q.latch.CanStop() {
		return ErrCannotStop
	}
	q.latch.WaitStopped() // wait for the dispatch loop to exit
	defer q.latch.Reset() // reset the latch in case we have to start again

	timeoutContext, cancel := context.WithTimeout(q.Background(), q.ShutdownGracePeriodOrDefault())
	defer cancel()

	if remainingWork := len(q.Work); remainingWork > 0 {
		var workItem T
		for x := 0; x < remainingWork; x++ {
			// check the timeout first
			select {
			case <-timeoutContext.Done():
				return nil
			default:
			}

			select {
			case <-timeoutContext.Done():
				return nil
			case workItem = <-q.Work:
				select {
				case <-timeoutContext.Done():
					return nil
				case worker := <-q.availableWorkers:
					select {
					case <-timeoutContext.Done():
						return nil
					case worker.Work <- workItem:
					}
				}
			}
		}
	}

	workersStopped := make(chan struct{})
	go func() {
		defer close(workersStopped)
		wg := sync.WaitGroup{}
		wg.Add(len(q.workers))
		for _, worker := range q.workers {
			go func(w *Worker[T]) {
				defer wg.Done()
				w.StopContext(timeoutContext)
			}(worker)
		}
		wg.Wait()
	}()

	select {
	case <-timeoutContext.Done():
		return nil
	case <-workersStopped:
		return nil
	}
}

// Close stops the queue.
// Any work left in the queue will be discarded.
func (q *Queue[T]) Close() error {
	q.latch.WaitStopped()
	q.latch.Reset()
	return nil
}

func (q *Queue[T]) ensureLatch() {
	q.latchMu.Lock()
	defer q.latchMu.Unlock()
	if q.latch == nil {
		q.latch = NewLatch()
	}
}

// ReturnWorker returns a given worker to the worker queue.
func (q *Queue[T]) returnWorker(ctx context.Context, worker *Worker[T]) error {
	select {
	case <-ctx.Done():
		return nil
	case q.availableWorkers <- worker:
		return nil
	}
}
