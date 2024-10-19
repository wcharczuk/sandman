package apputil

import (
	"context"
	"flag"
	"log/slog"
	"strings"

	"go.charczuk.com/sdk/cliutil"
	"go.charczuk.com/sdk/configutil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/log"
)

// DBEntryPointConfigProvider is a type that can provide
// an entrypoint config.
type DBEntryPointConfigProvider interface {
	LoggerProvider
	DBProvider
	MetaProvider
}

// EntryPoint is a top level handler for a database backed app.
//
// It handles reading the config, setting up the logger,
// opening the database connection, and based on comandline
// flags handling calling specific handlers for initializing the
// database and applying migrations on start.
type DBEntryPoint[T DBEntryPointConfigProvider] struct {
	Setup   func(context.Context, T) error
	Migrate func(context.Context, T, *db.Connection) error
	Start   func(context.Context, T, *db.Connection) error

	config              T
	flagDatabaseSetup   bool
	flagDatabaseMigrate bool
	flagStart           bool
	didInit             bool
}

// Init should be run before `Run` and registers flags and the like.
func (e *DBEntryPoint[T]) Init() {
	if e.didInit {
		return
	}
	e.didInit = true
	flag.BoolVar(&e.flagDatabaseSetup, "db-setup", false, "if we should run the first time setup for the database")
	flag.BoolVar(&e.flagDatabaseMigrate, "db-migrate", false, "if we should apply database migrations")
	flag.BoolVar(&e.flagStart, "start", true, "if we should start the server (false will exit after other steps complete)")
}

// Main is the actual function that needs to be called in Main.
func (e *DBEntryPoint[T]) Main() {
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
		logger.Info("using config path(s)", slog.String("config_paths", strings.Join(configPaths, ", ")))
	} else {
		logger.Info("using environment resolved config")
	}
	if e.flagDatabaseSetup && e.Setup != nil {
		logger.Info("running first time database setup", slog.String("database", e.config.GetDB().Database))
		if err := e.Setup(ctx, e.config); err != nil {
			cliutil.Fatal(err)
		}
		logger.Info("running first time database setup complete")
	} else {
		logger.Debug("skipping running first time database setup")
	}
	conn, err := db.New(
		db.OptConfig(e.config.GetDB()),
		db.OptLog(logger),
	)
	if err != nil {
		cliutil.Fatal(err)
	}
	if err = conn.Open(); err != nil {
		cliutil.Fatal(err)
	}
	defer conn.Close()

	if e.flagDatabaseMigrate && e.Migrate != nil {
		logger.Info("applying database migrations")
		if err := e.Migrate(ctx, e.config, conn); err != nil {
			cliutil.Fatal(err)
		}
		logger.Info("applying database migrations complete")
	} else {
		logger.Debug("skipping database migrations")
	}

	if e.config.GetMeta().IsProdlike() {
		logger.Debug("using database", slog.String("dsn", conn.CreateLoggingDSN()))
	}

	if e.flagStart && e.Start != nil {
		if err := e.Start(ctx, e.config, conn); err != nil {
			cliutil.Fatal(err)
		}
	}
}
