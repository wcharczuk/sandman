package async

import (
	"context"
	"fmt"
	"sync"
)

// BatchContext returns a new Batch and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
func BatchContext(ctx context.Context) (*Batch, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Batch{cancel: cancel}, ctx
}

// Batch is a collection of goroutines working on subtasks that are part of
// the same overall task.
type Batch struct {
	cancel  func()
	wg      sync.WaitGroup
	sem     chan struct{}
	errOnce sync.Once
	err     error
}

func (b *Batch) done() {
	if b.sem != nil {
		<-b.sem
	}
	b.wg.Done()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (b *Batch) Wait() error {
	b.wg.Wait()
	if b.cancel != nil {
		b.cancel()
	}
	return b.err
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (b *Batch) Go(f func() error) {
	if b.sem != nil {
		b.sem <- struct{}{}
	}

	b.wg.Add(1)
	go func() {
		defer b.done()

		if err := f(); err != nil {
			b.errOnce.Do(func() {
				b.err = err
				if b.cancel != nil {
					b.cancel()
				}
			})
		}
	}()
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (b *Batch) TryGo(f func() error) bool {
	if b.sem != nil {
		select {
		case b.sem <- struct{}{}:
			// Note: this allows barging iff channels in general allow barging.
		default:
			return false
		}
	}

	b.wg.Add(1)
	go func() {
		defer b.done()

		if err := f(); err != nil {
			b.errOnce.Do(func() {
				b.err = err
				if b.cancel != nil {
					b.cancel()
				}
			})
		}
	}()
	return true
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (b *Batch) SetLimit(n int) {
	if n < 0 {
		b.sem = nil
		return
	}
	if len(b.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(b.sem)))
	}
	b.sem = make(chan struct{}, n)
}
