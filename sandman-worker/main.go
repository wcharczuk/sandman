package main

import (
	"context"
	"expvar"
	"flag"
	"net/http"
	"os"
	"sandman/pkg/config"
	"sandman/pkg/model"
	"sandman/pkg/worker"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/configutil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/db/migration"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"
)

var (
	flagExpvarBindAddr = flag.String("expvar-bind-addr", "", "The expvar bind address")
	flagHostname       = flag.String("hostname", "", "The worker hostname")
	flagOrdinal        = flag.Int("ordinal", 0, "The ordinal of the worker (or index in a worker ring")
)

type workerConfig struct {
	config.Config `yaml:",inline"`

	ExpvarBindAddr string
}

func (wc *workerConfig) Resolve(ctx context.Context) error {
	return configutil.Resolve(
		ctx,
		(&wc.Config).Resolve,
		configutil.Set(&wc.Hostname, configutil.Lazy(flagHostname), configutil.Env[string]("HOSTNAME")),
		configutil.Set(&wc.ExpvarBindAddr, configutil.Lazy(flagExpvarBindAddr), configutil.Env[string]("EXPVAR_BIND_ADDR"), configutil.Const(":8090")),
	)
}

var entrypoint = apputil.DBEntryPoint[workerConfig]{
	Setup: func(ctx context.Context, cfg workerConfig) error {
		return nil
	},
	Migrate: func(ctx context.Context, cfg workerConfig, dbc *db.Connection) error {
		return model.Migrations(
			migration.OptLog(log.GetLogger(ctx)),
		).Apply(ctx, dbc)
	},
	Start: func(ctx context.Context, cfg workerConfig, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-worker")
		modelMgr := &model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}
		w := worker.New(cfg.Hostname, modelMgr)
		w.Vars().Publish()
		go func() {
			if err := http.ListenAndServe(cfg.ExpvarBindAddr, expvar.Handler()); err != nil {
				log.GetLogger(ctx).Error("expvar server error", log.Any("err", err))
			}
		}()
		return w.Run(ctx)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
