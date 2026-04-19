package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"go.charczuk.com/sdk/cliutil"
	"go.charczuk.com/sdk/graceful"
	"go.charczuk.com/sdk/supervisor"
)

var services cliutil.FlagStrings

func main() {
	flag.Var(&services, "service", "Services to start (defaults to all if omitted, can be multiple")
	flag.Parse()

	s := &supervisor.Supervisor{
		Services: []*supervisor.Service{
			{
				Name:       "sandman-srv",
				Enabled:    services.HasOrUnset("sandman-srv"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-srv/main.go"},
				Env:        os.Environ(),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-srv/...",
				},
				// Stdout:        supervisor.PrefixWriter{Prefix: "sandman-srv| ", Writer: os.Stderr},
				// Stderr:        supervisor.PrefixWriter{Prefix: "sandman-srv| ", Writer: os.Stderr},
				RestartPolicy: new(restartPolicy),
			},
			{
				Name:       "sandman-control",
				Enabled:    services.HasOrUnset("sandman-control"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args: []string{"run", "sandman-control/main.go",
					"--mode=fork",
				},
				Env: os.Environ(),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-control/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "sandman-control| ", Writer: os.Stderr},
				Stderr:        supervisor.PrefixWriter{Prefix: "sandman-control| ", Writer: os.Stderr},
				RestartPolicy: new(restartPolicy),
			},
			{
				Name:       "hook-target",
				Enabled:    services.HasOrUnset("hook-target"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args: []string{"run", "scripts/hook-target/main.go",
					"--skip-logs",
					"--listen-addr=:8080",
				},
				Env: os.Environ(),
				WatchedPaths: []string{
					"./scripts/hook-target/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "hook-target| ", Writer: os.Stderr},
				Stderr:        supervisor.PrefixWriter{Prefix: "hook-target| ", Writer: os.Stderr},
				RestartPolicy: new(restartPolicy),
			},
		},
	}
	fmt.Printf("!! nodes dev starting\n")
	fmt.Printf("!! you can force a restart of the sub-process tree with: kill -HUP %d\n", os.Getpid())
	if err := graceful.StartForShutdown(context.Background(), s); err != nil {
		cliutil.MaybeFatal(err)
	}
}

var (
	_ supervisor.RestartPolicy = (*restartPolicy)(nil)
)

type restartPolicy struct{}

func (rp restartPolicy) ShouldRestart(_ context.Context, _ *supervisor.ServiceHistory) bool {
	return true
}

func (rp restartPolicy) Delay(_ context.Context, history *supervisor.ServiceHistory) time.Duration {
	numberOfRecentFailures := time.Duration(len(history.RecentFailures()))
	return numberOfRecentFailures * time.Second
}
