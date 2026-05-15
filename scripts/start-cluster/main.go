package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sandman/pkg/supervisor"
)

const shutdownGrace = 15 * time.Second

func main() {
	mikoshi, err := mikoshiPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	nodes := []struct {
		name string
		addr string
	}{
		{"node-1", "127.0.0.1:26257"},
		{"node-2", "127.0.0.1:26258"},
		{"node-3", "127.0.0.1:26259"},
	}
	join := "127.0.0.1:26257,127.0.0.1:26258,127.0.0.1:26259"

	var (
		wg   sync.WaitGroup
		cmds []*exec.Cmd
	)
	for _, n := range nodes {
		cmd := exec.Command(mikoshi,
			"start",
			"--insecure",
			"--listen-addr="+n.addr,
			"--join="+join,
			"--store=type=mem,size=2GiB",
		)
		cmd.Env = os.Environ()
		cmd.Stdout = supervisor.PrefixWriter{Prefix: n.name + "| ", Writer: os.Stdout}
		cmd.Stderr = supervisor.PrefixWriter{Prefix: n.name + "| ", Writer: os.Stderr}
		// Put each child in its own process group so we can signal the whole
		// group on shutdown and so signals delivered to our terminal foreground
		// group (e.g. Ctrl+C) don't race us to the children.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to start %s: %+v\n", n.name, err)
			cancel()
			shutdown(cmds, &wg)
			os.Exit(1)
		}
		cmds = append(cmds, cmd)
		wg.Add(1)
		go func(c *exec.Cmd) {
			defer wg.Done()
			_ = c.Wait()
		}(cmd)
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdown(cmds, &wg)
	}()

	if err := awaitListen(ctx, "127.0.0.1:26257"); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		cancel()
		<-shutdownDone
		os.Exit(1)
	}

	mikoshiInit := exec.CommandContext(ctx, mikoshi,
		"init",
		"--insecure",
		"--host=127.0.0.1:26257",
	)
	mikoshiInit.Stdout = supervisor.PrefixWriter{Prefix: "init| ", Writer: os.Stdout}
	mikoshiInit.Stderr = supervisor.PrefixWriter{Prefix: "init| ", Writer: os.Stderr}

	if err := mikoshiInit.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		cancel()
		<-shutdownDone
		os.Exit(1)
	}

	wg.Wait()
	cancel()
	<-shutdownDone
}

// shutdown sends SIGTERM to each child's process group, waits up to
// shutdownGrace for the wait group to drain, then SIGKILLs anything still
// running so we don't leave orphans behind.
func shutdown(cmds []*exec.Cmd, wg *sync.WaitGroup) {
	for _, c := range cmds {
		if c.Process == nil {
			continue
		}
		_ = syscall.Kill(-c.Process.Pid, syscall.SIGTERM)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(shutdownGrace):
		for _, c := range cmds {
			if c.Process == nil {
				continue
			}
			_ = syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
		}
		<-done
	}
}

func awaitListen(ctx context.Context, addr string) error {
	start := time.Now()
	var err error
	for time.Since(start) < 30*time.Second {
		if err = isListening(ctx, addr); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for listen; last error: %v", err)
}

func isListening(ctx context.Context, addr string) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func mikoshiPath() (string, error) {
	return exec.LookPath("mikoshi")
}
