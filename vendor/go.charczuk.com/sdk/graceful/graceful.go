package graceful

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"go.charczuk.com/sdk/errutil"
)

// Graceful is the main entrypoint for hosting graceful processes.
type Graceful struct {
	Hosted              []Service
	ShutdownSignal      chan os.Signal
	RestartSignal       chan os.Signal
	ShutdownGracePeriod time.Duration
	Log                 Logger
}

// StartForShutdown starts, and prepares to gracefully stop, a set hosted
// processes based on a provided context's cancellation.
//
// The context is used to stop the goroutines this function spawns,
// as well as call `Stop(...)` on the hosted processes when it cancels.
func (g Graceful) StartForShutdown(ctx context.Context) error {
	shouldShutdown := make(chan struct{})
	shouldRestart := make(chan struct{})
	serverExited := make(chan struct{})

	if g.ShutdownSignal != nil {
		go func() {
			MaybeDebugf(g.Log, "graceful background; waiting for shutdown signal")
			select {
			case <-ctx.Done():
				return
			case <-g.ShutdownSignal:
				MaybeDebugf(g.Log, "graceful background; shutdown signal received, canceling context")
				_ = safelyClose(shouldShutdown)
			}
		}()
	}

	if g.RestartSignal != nil {
		go func() {
			for {
				MaybeDebugf(g.Log, "graceful background; waiting for restart signal")
				select {
				case <-ctx.Done():
					return
				case <-g.RestartSignal:
					// NOTE(wc): we don't close here because we may need to do this more than once!
					shouldRestart <- struct{}{}
					MaybeDebugf(g.Log, "graceful background; shutdown signal received, canceling context")
				}
			}
		}()
	}

	var waitShutdownComplete sync.WaitGroup
	waitShutdownComplete.Add(len(g.Hosted))

	var waitServerExited sync.WaitGroup
	waitServerExited.Add(len(g.Hosted))

	hostedErrors := make(chan error, 2*len(g.Hosted))

	for _, hostedInstance := range g.Hosted {
		// start the instance
		go func(instance Service) {
			defer func() {
				_ = safelyClose(serverExited)
				waitServerExited.Done() // signal the normal exit process is done
			}()
			if err := instance.Start(ctx); err != nil {
				hostedErrors <- err
			}
		}(hostedInstance)

		// wait to restart the instance
		go func(instance Service) {
			<-shouldRestart
			if err := instance.Restart(ctx); err != nil {
				hostedErrors <- err
			}
		}(hostedInstance)

		// wait to stop the instance
		go func(instance Service) {
			defer waitShutdownComplete.Done()
			<-shouldShutdown // the hosted process has been told to stop "gracefully"
			if err := instance.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
				hostedErrors <- err
			}
		}(hostedInstance)
	}

	select {
	case <-ctx.Done(): // if we've issued a shutdown, wait for the server to exit
		_ = safelyClose(shouldShutdown)
		g.gracefulShutdown(ctx, &waitShutdownComplete, &waitServerExited)
	case <-serverExited:
		// if any of the servers exited on their
		// own, we should crash the whole party
		_ = safelyClose(shouldShutdown)
		waitShutdownComplete.Wait()
		g.gracefulExit(ctx, &waitShutdownComplete)
	}
	if errorCount := len(hostedErrors); errorCount > 0 {
		var err error
		for x := 0; x < errorCount; x++ {
			err = errutil.AppendFlat(err, <-hostedErrors)
		}
		return err
	}
	return nil
}

func (g Graceful) gracefulShutdown(ctx context.Context, waitShutdownComplete, waitServerExited *sync.WaitGroup) {
	if g.ShutdownGracePeriod > 0 {
		didShutdown := make(chan struct{})
		go func() {
			defer safelyClose(didShutdown)
			waitShutdownComplete.Wait()
			waitServerExited.Wait()
		}()

		mustShutdownTimer := time.NewTimer(g.ShutdownGracePeriod)
		defer mustShutdownTimer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-mustShutdownTimer.C:
			return
		case <-didShutdown:
			return
		}
	}
	waitShutdownComplete.Wait()
	waitServerExited.Wait()
}

func (g Graceful) gracefulExit(ctx context.Context, waitShutdownComplete *sync.WaitGroup) {
	if g.ShutdownGracePeriod > 0 {
		didShutdown := make(chan struct{})
		go func() {
			defer safelyClose(didShutdown)
			waitShutdownComplete.Wait()
		}()
		mustShutdownTimer := time.NewTimer(g.ShutdownGracePeriod)
		defer mustShutdownTimer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-mustShutdownTimer.C:
			return
		case <-didShutdown:
			return
		}
	}
	waitShutdownComplete.Wait()
}

func safelyClose(c chan struct{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errutil.New(r)
		}
	}()
	close(c)
	return
}

func safely(action func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errutil.New(r)
		}
	}()
	action()
	return
}
