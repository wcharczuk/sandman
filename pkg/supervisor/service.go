package supervisor

import (
	"context"
	"fmt"
	"io"
	"sync"
	"syscall"
	"time"

	"github.com/rjeczalik/notify"
)

// Service is a specific program to start.
type Service struct {
	Name                      string
	Enabled                   bool
	Background                func(context.Context) context.Context
	Command                   string
	Args                      []string
	ShutdownSignal            syscall.Signal
	Env                       []string
	WorkDir                   string
	WatchedPaths              []string
	WatchedPathChangeDebounce time.Duration
	Stdin                     io.Reader
	Stdout                    io.Writer
	Stderr                    io.Writer
	RestartPolicy             RestartPolicy
	OnStart                   []func(context.Context)
	OnRestart                 []func(context.Context)
	OnExit                    []func(context.Context)
	SubprocessProvider        SubprocessProvider // used for testing
	FileEventProvider         FileEventProvider  // used for testing

	// internal fields
	// no peeking!

	subprocessMu           sync.Mutex
	subprocess             Subprocess
	fsevents               chan notify.EventInfo
	fileEventDebouncedAtMu sync.Mutex
	fileEventDebouncedAt   time.Time
	history                ServiceHistory
	crashed                func(error)
	finalizer              func()
	stopping               bool
	restarting             bool
	done                   chan struct{}
}

// Start starts the service.
func (s *Service) Start(ctx context.Context) error {
	if !s.Enabled {
		return nil
	}

	if s.safeSubprocessIsSet() {
		return nil
	}
	if err := s.safeInitializeExecHandle(ctx); err != nil {
		return err
	}

	s.history.StartedAt = time.Now()
	s.done = make(chan struct{})

	if len(s.WatchedPaths) > 0 {
		s.fsevents = make(chan notify.EventInfo, 1)
		notifyProvider := s.fileEventProviderOrDefault()
		for _, watchedPath := range s.WatchedPaths {
			if err := notifyProvider.Notify(watchedPath, s.fsevents); err != nil {
				return err
			}
		}
	}
	if err := s.subprocess.Start(); err != nil {
		return err
	}
	for _, handler := range s.OnStart {
		handler(ctx)
	}

	if len(s.WatchedPaths) > 0 {
		go func() {
			var e notify.EventInfo
			var restartErr error
			for {
				select {
				case <-s.done:
					return
				case e = <-s.fsevents:
					s.errPrintf("restarting on filesystem changes")
					restartErr = s.safeDebouncedSignalOnWatchedEvent(e)
					if restartErr != nil {
						s.errPrintf("restarting on filesystem changes; error on terminate signal; %v", restartErr)
					}
				}
			}
		}()
	}

	// fork the goroutine which will handle the process itself, including restarts and termination.
	go func() {
		// finalErr is the error that will be passed to the crashed handler.
		var finalErr error

		defer func() {
			// do not call the crashed handler if we're
			// specifically being told to stop!
			if !s.stopping && finalErr != nil && s.crashed != nil {
				s.crashed(finalErr)
			}
			if s.finalizer != nil {
				s.finalizer()
			}
			for _, handler := range s.OnExit {
				handler(ctx)
			}
			close(s.done)
		}()
		var waitErr, startErr error
		var delay time.Duration
		for {
			s.errPrintf("started with pid: %v", s.subprocess.Pid())

			// wait for the sub-process to exit
			//
			// waitErr will be set if the process was terminated with a signal!
			waitErr = s.subprocess.Wait()

			// add the event to history but factor that the restart
			// may have been by a file change that we _do not_ want to
			// record as a failure!
			s.addHistoryEvent(waitErr)

			// we should only consider restarting if we are _not_ stopping
			if s.maybeShouldRestart(ctx) {

				// we may need to delay the restart
				if delay = s.maybeShouldDelayRestart(ctx); delay > 0 {
					s.errPrintf("delaying %v to restart", delay.Round(time.Millisecond))
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return
					}
				}

				if s.stopping {
					s.errPrintf("exiting on shutdown")
					finalErr = waitErr
					return
				} else if s.restarting {
					s.errPrintf("starting after restart")
				} else if waitErr != nil {
					s.errPrintf("starting after process exit error: %v", waitErr)
				} else {
					s.errPrintf("starting after process exit")
				}

				// re-initialize the sub-process
				s.safeInitializeExecHandle(ctx)

				// call the on restart handlers (before we actually start!)
				// but _after_ we've re-initialized the sub-process
				for _, handler := range s.OnRestart {
					handler(ctx)
				}

				if startErr = s.subprocess.Start(); startErr != nil {
					s.errPrintf("failed to restart")
					finalErr = startErr
					return
				}

			} else {
				if s.stopping {
					s.errPrintf("exiting on shutdown")
				} else {
					s.errPrintf("exiting based on exhausting restart policy")
				}
				finalErr = waitErr
				return
			}
		}
	}()
	return nil
}

// Stop stops the service.
func (s *Service) Stop() error {
	s.subprocessMu.Lock()
	defer s.subprocessMu.Unlock()
	if s.subprocess != nil {
		s.stopping = true
		notify.Stop(s.fsevents)
		return s.signalTerminate()
	}
	return nil
}

// Restart tells the service to quit with the shutdown signal restarting the serivce.
func (s *Service) Restart() (err error) {
	s.subprocessMu.Lock()
	defer s.subprocessMu.Unlock()
	s.restarting = true
	err = s.signalTerminate()
	return
}

//
// internal methods
//

func (s *Service) watchedPathChangeDebounceOrDefault() time.Duration {
	if s.WatchedPathChangeDebounce > 0 {
		return s.WatchedPathChangeDebounce
	}
	return 500 * time.Millisecond
}

func (s *Service) safeDebouncedSignalOnWatchedEvent(_ notify.EventInfo) error {
	s.fileEventDebouncedAtMu.Lock()
	defer s.fileEventDebouncedAtMu.Unlock()

	if s.fileEventDebouncedAt.IsZero() || time.Since(s.fileEventDebouncedAt) > s.watchedPathChangeDebounceOrDefault() {
		s.fileEventDebouncedAt = time.Now()
		return s.Restart()
	}
	return nil
}

func (s *Service) signalTerminate() (err error) {
	if s.subprocess == nil {
		return
	}
	var signal syscall.Signal
	if s.ShutdownSignal > 0 {
		signal = s.ShutdownSignal
	} else {
		signal = syscall.SIGINT
	}
	s.errPrintf("being sent terminate signal: %v", signal)
	err = s.subprocess.Signal(signal)
	return
}

func (s *Service) maybeShouldRestart(ctx context.Context) bool {
	// we _never_ restart if we're stopping.
	if s.stopping {
		return false
	}

	// we _always_ restart if it's because of an explicit
	// restart or a filesystem change.
	if s.restarting {
		return true
	}

	// return the result of the restart policy.
	return s.RestartPolicy != nil && s.RestartPolicy.ShouldRestart(ctx, &s.history)
}

func (s *Service) maybeShouldDelayRestart(ctx context.Context) time.Duration {
	if s.RestartPolicy != nil {
		return s.RestartPolicy.Delay(ctx, &s.history)
	}
	return 0
}

func (s *Service) safeSubprocessIsSet() (set bool) {
	s.subprocessMu.Lock()
	set = s.subprocess != nil
	s.subprocessMu.Unlock()
	return
}

func (s *Service) safeInitializeExecHandle(ctx context.Context) error {
	s.subprocessMu.Lock()
	defer s.subprocessMu.Unlock()
	s.stopping = false
	s.restarting = false
	if s.Background != nil {
		ctx = s.Background(ctx)
	}
	sub, err := s.subprocessProviderOrDefault().Exec(ctx, s)
	if err != nil {
		return err
	}
	s.subprocess = sub
	return nil
}

func (s *Service) fileEventProviderOrDefault() FileEventProvider {
	if s.FileEventProvider != nil {
		return s.FileEventProvider
	}
	return new(NotifyProvider)
}

func (s *Service) subprocessProviderOrDefault() SubprocessProvider {
	if s.SubprocessProvider != nil {
		return s.SubprocessProvider
	}
	return new(ExecSubprocessProvider)
}

func (s *Service) addHistoryEvent(err error) {
	now := time.Now()

	// elide the error on restart as we do _not_ consider
	// signal errors from restarts as real failures
	// for restart policies.
	if s.restarting {
		s.history.Exits = append(s.history.Exits, Exit{
			Timestamp: now,
		})
		return
	}

	s.history.Exits = append(s.history.Exits, Exit{
		Timestamp: now,
		Error:     err,
	})
}

func (s *Service) errPrintf(format string, args ...any) {
	if s.Stderr != nil {
		fmt.Fprintf(s.Stderr, "[supervisor] %s process %s\n", s.Name, fmt.Sprintf(format, args...))
	}
}
