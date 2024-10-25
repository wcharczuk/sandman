package supervisor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"go.charczuk.com/sdk/errutil"
	"go.charczuk.com/sdk/graceful"
)

var _ graceful.Service = (*Supervisor)(nil)

// Supervisor is a collection of services that should be started / restarted.
type Supervisor struct {
	Services []*Service

	status  int32
	crashed chan error
	waits   sync.WaitGroup
}

// StatusTypes
const (
	StatusStopped  int32 = iota
	StatusStarting int32 = iota
	StatusRunning  int32 = iota
	StatusStopping int32 = iota
)

// Start starts the services and blocks.
func (s *Supervisor) Start(ctx context.Context) error {
	if err := s.StartAsync(ctx); err != nil {
		return err
	}
	return s.Wait(ctx)
}

// Wait blocks until the services exit.
func (s *Supervisor) Wait(ctx context.Context) error {
	if atomic.LoadInt32(&s.status) != StatusRunning {
		return nil
	}
	defer func() {
		atomic.StoreInt32(&s.status, StatusStopped)
	}()

	done := make(chan struct{})
	go func() {
		s.waits.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-s.crashed:
		s.status = StatusStopping
		for _, service := range s.Services {
			_ = service.Stop()
		}
		return err
	case <-done:
		return nil
	}
}

// StartAsync starts the supervisor and does not block.
func (s *Supervisor) StartAsync(ctx context.Context) (err error) {
	if !atomic.CompareAndSwapInt32(&s.status, StatusStopped, StatusStarting) {
		return
	}
	defer func() {
		if err != nil {
			atomic.StoreInt32(&s.status, StatusStopped)
		} else {
			atomic.StoreInt32(&s.status, StatusRunning)
		}
	}()

	s.waits = sync.WaitGroup{}
	s.crashed = make(chan error, len(s.Services))
	for x := 0; x < len(s.Services); x++ {
		s.Services[x].crashed = func(err error) {
			s.crashed <- err
		}
		s.Services[x].finalizer = func() {
			s.waits.Done()
		}

		// skip disabled services
		if !s.Services[x].Enabled {
			fmt.Println("skipping starting service", s.Services[x].Name)
			continue
		}

		if err = s.Services[x].Start(ctx); err != nil {
			for y := 0; y < x; y++ {
				_ = s.Services[y].Stop()
			}
			return
		}
		s.waits.Add(1)
	}
	return
}

// Restart restarts all the services.
//
// If there are errors on restart those errors are returned but
// no recovery or coordinated shutdown attempt is made.
func (s *Supervisor) Restart(_ context.Context) (err error) {
	if !atomic.CompareAndSwapInt32(&s.status, StatusRunning, StatusStopping) {
		return
	}
	defer func() {
		atomic.StoreInt32(&s.status, StatusRunning)
	}()
	for _, service := range s.Services {
		if serviceErr := service.Restart(); serviceErr != nil {
			err = errutil.AppendFlat(err, serviceErr)
		}
	}
	return
}

// Stop stops the supervisor, implementing graceful.
func (s *Supervisor) Stop(_ context.Context) (err error) {
	if !atomic.CompareAndSwapInt32(&s.status, StatusRunning, StatusStopping) {
		return
	}
	for _, service := range s.Services {
		if serviceErr := service.Stop(); serviceErr != nil {
			err = errutil.AppendFlat(err, serviceErr)
		}
	}
	return
}
