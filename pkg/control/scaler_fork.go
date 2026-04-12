package control

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/supervisor"
)

// ForkScaler manages worker replicas as forked `go run` processes.
// This is intended for local development use only.
type ForkScaler struct {
	mu      sync.Mutex
	workers []*managedProcess
}

type managedProcess struct {
	cmd      *exec.Cmd
	hostname string
	done     chan struct{}
}

func (s *ForkScaler) SetDesiredScale(ctx context.Context, desired int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := int32(len(s.workers))
	logger := log.GetLogger(ctx)

	if desired == current {
		return nil
	}

	if desired > current {
		for i := current; i < desired; i++ {
			hostname := fmt.Sprintf("worker-%d", i)
			cmd := exec.Command("go", "run", "./sandman-worker",
				"--hostname", hostname,
			)
			cmd.Stdout = supervisor.PrefixWriter{Prefix: fmt.Sprintf("%s| ", hostname), Writer: os.Stderr}
			cmd.Stderr = supervisor.PrefixWriter{Prefix: fmt.Sprintf("%s| ", hostname), Writer: os.Stderr}
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start worker %d: %w", i, err)
			}
			w := &managedProcess{
				cmd:      cmd,
				hostname: hostname,
				done:     make(chan struct{}),
			}
			go func() {
				cmd.Wait()
				close(w.done)
			}()
			logger.Info("fork scaler; started worker",
				log.String("hostname", hostname),
				log.Int("pid", cmd.Process.Pid),
			)
			s.workers = append(s.workers, w)
		}
	} else {
		for i := current - 1; i >= desired; i-- {
			w := s.workers[i]
			logger.Info("fork scaler; stopping worker",
				log.String("hostname", w.hostname),
			)
			s.stopProcess(w)
		}
		s.workers = s.workers[:desired]
	}
	return nil
}

// Close stops all managed worker processes.
func (s *ForkScaler) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, w := range s.workers {
		s.stopProcess(w)
	}
	s.workers = nil
}

func (s *ForkScaler) stopProcess(w *managedProcess) {
	if w.cmd.Process == nil {
		return
	}
	_ = w.cmd.Process.Signal(os.Interrupt)
	select {
	case <-w.done:
	case <-time.After(10 * time.Second):
		_ = w.cmd.Process.Kill()
		<-w.done
	}
}
