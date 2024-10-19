package apputil

import (
	"context"
	"flag"
	"strings"

	"go.charczuk.com/sdk/cliutil"
	"go.charczuk.com/sdk/configutil"
	"go.charczuk.com/sdk/log"
)

// EntryPointConfigProvider is a type that can provide
// an entrypoint config.
type EntryPointConfigProvider interface {
	LoggerProvider
}

// EntryPoint is a top level handler for a simple app.
//
// It handles reading the config, and setting up the logger.
type EntryPoint[T EntryPointConfigProvider] struct {
	Setup func(context.Context, T) error
	Start func(context.Context, T) error

	config    T
	flagSetup bool
	flagStart bool
	didInit   bool
}

// Init should be run before `Run` and registers flags and the like.
func (e *EntryPoint[T]) Init() {
	if e.didInit {
		return
	}
	e.didInit = true
	flag.BoolVar(&e.flagSetup, "setup", false, "if we should run the first time setup steps")
	flag.BoolVar(&e.flagStart, "start", true, "if we should start the server (false will exit after other steps complete)")
}

// Main is the actual function that needs to be called in Main.
func (e *EntryPoint[T]) Main() {
	if !e.didInit {
		e.Init()
	}
	flag.Parse()
	ctx := context.Background()
	configPaths := configutil.MustRead(&e.config)
	logger := log.New(
		log.OptConfig(e.config.GetLogger()),
	)

	ctx = log.WithLogger(ctx, logger)
	if len(configPaths) > 0 {
		logger.Info("using config path(s)", "config_paths", strings.Join(configPaths, ", "))
	} else {
		logger.Info("using environment resolved config")
	}
	if e.flagSetup && e.Setup != nil {
		logger.Info("running first time setup")
		if err := e.Setup(ctx, e.config); err != nil {
			cliutil.Fatal(err)
		}
		logger.Info("running first time setup complete")
	} else {
		logger.Debug("skipping running first time setup")
	}
	if e.flagStart && e.Start != nil {
		logger.Info("starting")
		if err := e.Start(ctx, e.config); err != nil {
			cliutil.Fatal(err)
		}
	} else {
		logger.Info("exiting")
	}
}
