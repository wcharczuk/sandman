package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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
				Name:       "worker-00",
				Enabled:    services.HasOrUnset("workers"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-worker/main.go", "--expvar-bind-addr=:8090", "--hostname=worker-00"},
				Env:        os.Environ(),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-worker/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "worker-00-out| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "worker-00-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
			{
				Name:       "worker-01",
				Enabled:    services.HasOrUnset("workers"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-worker/main.go", "--expvar-bind-addr=:8091", "--hostname=worker-01"},
				Env:        os.Environ(),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-worker/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "worker-01-out| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "worker-01-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
			{
				Name:       "sandman-srv",
				Enabled:    services.HasOrUnset("sandman-srv"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-srv/main.go"},
				Env:        append([]string{"EXPVAR_BIND_ADDR=:8081"}, os.Environ()...),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-srv/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "server| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "server-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
			{
				Name:       "target",
				Enabled:    services.HasOrUnset("target"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "scripts/target/main.go"},
				Env:        append([]string{"BIND_ADDR=:8080"}, os.Environ()...),
				WatchedPaths: []string{
					"./scripts/target/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "target| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "target-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
		},
	}
	fmt.Printf("!! nodes dev starting\n")
	fmt.Printf("!! you can force a restart of the sub-process tree with: kill -HUP %d\n", os.Getpid())
	if err := graceful.StartForShutdown(context.Background(), s); err != nil {
		cliutil.Fatal(err)
	}
}
