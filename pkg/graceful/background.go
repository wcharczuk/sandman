package graceful

import (
	"context"
	"os"
	"os/signal"
)

// OptBackgroundSignals sets the signals.
func OptBackgroundSignals(signals ...os.Signal) BackgroundOption {
	return func(bo *BackgroundOptions) { bo.Signals = signals }
}

// OptBackgroundLog sets the logger.
func OptBackgroundLog(log Logger) BackgroundOption {
	return func(bo *BackgroundOptions) { bo.Log = log }
}

// OptBackgroundSkipStopOnSignal sets if we should stop the signal channel on stop.
func OptBackgroundContext(ctx context.Context) BackgroundOption {
	return func(bo *BackgroundOptions) { bo.Context = ctx }
}

// BackgroundOption mutates background options
type BackgroundOption func(*BackgroundOptions)

// BackgroundOptions are options for the background context.
type BackgroundOptions struct {
	// Context is a root context, if unset `context.Background()` is used.
	Context context.Context
	// Signals are the specific os signals to intercept.
	Signals []os.Signal
	// Log holds an reference to a graceful logger.
	Log Logger
}

// Background yields a context that will signal `<-ctx.Done()` when
// a signal is sent to the process (as specified in `DefaultShutdownSignals`).
//
// This context will cancel only (1) time.
func Background(opts ...BackgroundOption) context.Context {
	options := BackgroundOptions{
		Context: context.Background(),
		Signals: DefaultShutdownSignals,
	}
	for _, opt := range opts {
		opt(&options)
	}

	ctx, cancel := context.WithCancel(options.Context)
	shutdown := SignalNotify(options.Signals...)
	go func() {
		defer func() {
			signal.Stop(shutdown)
		}()
		MaybeDebugf(options.Log, "graceful background; waiting for shutdown signal")
		select {
		case <-options.Context.Done():
			return
		case <-shutdown:
			MaybeDebugf(options.Log, "graceful background; shutdown signal received, canceling context")
			cancel()
			return
		}

	}()
	return ctx
}
