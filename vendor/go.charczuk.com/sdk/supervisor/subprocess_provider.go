package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

type SubprocessProvider interface {
	Exec(context.Context, *Service) (Subprocess, error)
}

// Subprocess is a forked process.
type Subprocess interface {
	Start() error
	Pid() int
	Signal(syscall.Signal) error
	Wait() error
}

// ExecSubprocessProvider is the exec subprocess provider.
type ExecSubprocessProvider struct{}

func (e ExecSubprocessProvider) Exec(ctx context.Context, svc *Service) (Subprocess, error) {
	esp := new(ExecSubprocess)
	commandResolved, err := exec.LookPath(svc.Command)
	if err != nil {
		return nil, err
	}
	esp.handle = exec.CommandContext(ctx, commandResolved, svc.Args...)

	var dir string
	if svc.WorkDir != "" {
		dir = filepath.Clean(svc.WorkDir)
	}
	// setting this SysProcAttr is required to be able to kill the process "group"
	// that is spawned from our main process if the user sends a signal.
	esp.handle.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	esp.handle.Dir = dir
	esp.handle.Env = svc.Env
	esp.handle.Stdin = svc.Stdin
	esp.handle.Stdout = svc.Stdout
	esp.handle.Stderr = svc.Stderr
	return esp, nil
}

// ExecSubprocess is a subprocess that is implemented using `os/exec`.
type ExecSubprocess struct {
	handle *exec.Cmd
}

// Start invokes the subprocess.
func (esp ExecSubprocess) Start() error {
	return esp.handle.Start()
}

// Pid returns the underlying pid.
func (esp ExecSubprocess) Pid() int {
	if esp.handle != nil && esp.handle.Process != nil {
		return esp.handle.Process.Pid
	}
	return 0
}

func (esp ExecSubprocess) Signal(sig syscall.Signal) error {
	if esp.handle != nil && esp.handle.Process != nil {
		if esp.handle.ProcessState != nil && esp.handle.ProcessState.Exited() {
			return nil
		}
		return syscall.Kill(-esp.handle.Process.Pid, sig)
	}
	return nil
}

// Wait blocks on the subprocess.
func (esp ExecSubprocess) Wait() error {
	return esp.handle.Wait()
}

// MockSubprocessProvider is a provider for subprocesses that returns
// a mocked subprocess.
type MockSubprocessProvider struct{}

// Exec returns a new mocked subprocess.
func (m MockSubprocessProvider) Exec(ctx context.Context, svc *Service) (Subprocess, error) {
	return &mockSubprocess{
		ctx: ctx,
		svc: svc,
	}, nil
}

// MockSubprocess is a mocked subprocess.
type MockSubprocess interface {
	Subprocess
	Exit(error)
}

type mockSubprocess struct {
	ctx  context.Context
	svc  *Service
	wait chan error
}

func (msp *mockSubprocess) Start() error {
	if msp.wait != nil {
		return nil
	}
	msp.wait = make(chan error)
	return nil
}

func (msp *mockSubprocess) Pid() int { return 123 }

func (msp *mockSubprocess) Signal(sig syscall.Signal) error {
	if msp.wait == nil {
		return fmt.Errorf("no such process")
	}
	msp.wait <- fmt.Errorf("process exit 1")
	return nil
}

func (msp *mockSubprocess) Exit(err error) {
	if msp.wait != nil {
		msp.wait <- err
	}
}

func (msp *mockSubprocess) Wait() error {
	select {
	case <-msp.ctx.Done():
		return context.Canceled
	case err := <-msp.wait:
		return err
	}
}
