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
				Name:       "scheduler",
				Enabled:    services.HasOrUnset("scheduler"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-scheduler/main.go"},
				Env:        append([]string{"EXPVAR_BIND_ADDR=:8081"}, os.Environ()...),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-scheduler/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "scheduler| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "scheduler-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
			{
				Name:       "worker",
				Enabled:    services.HasOrUnset("worker"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-worker/main.go"},
				Env:        append([]string{"EXPVAR_BIND_ADDR=:8082"}, os.Environ()...),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-worker/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "worker| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "worker-err| ", Writer: os.Stdout},
				RestartPolicy: supervisor.RestartPolicySuccessiveFailures(5),
			},
			{
				Name:       "sandman-srv",
				Enabled:    services.HasOrUnset("sandman-srv"),
				Background: func(_ context.Context) context.Context { return context.Background() },
				Command:    "go",
				Args:       []string{"run", "sandman-srv/main.go"},
				Env:        append([]string{"EXPVAR_BIND_ADDR=:8083"}, os.Environ()...),
				WatchedPaths: []string{
					"./pkg/...",
					"./sandman-srv/...",
				},
				Stdout:        supervisor.PrefixWriter{Prefix: "worker| ", Writer: os.Stdout},
				Stderr:        supervisor.PrefixWriter{Prefix: "worker-err| ", Writer: os.Stdout},
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
