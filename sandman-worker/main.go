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
	flagExpvarListenAddr = flag.String("expvar-listen-addr", "", "The expvar listen address")
	flagHostname         = flag.String("hostname", "", "The worker hostname")
)

type workerConfig struct {
	config.Config    `yaml:",inline"`
	ExpvarListenAddr string
}

func (wc *workerConfig) Resolve(ctx context.Context) error {
	return configutil.Resolve(
		ctx,
		(&wc.Config).Resolve,
		configutil.Set(&wc.Hostname, configutil.Lazy(flagHostname), configutil.Env[string]("HOSTNAME")),
		configutil.Set(&wc.ExpvarListenAddr, configutil.Lazy(flagExpvarListenAddr), configutil.Env[string]("EXPVAR_LISTEN_ADDR")),
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
		workerOpts := []worker.WorkerOption{
			worker.OptBatchSize(cfg.Worker.BatchSizeOrDefault()),
			worker.OptPollingInterval(cfg.Worker.PollingIntervalOrDefault()),
		}
		if cfg.Worker.PrefetchWindow > 0 {
			workerOpts = append(workerOpts, worker.OptPrefetchWindow(cfg.Worker.PrefetchWindow))
		}
		if cfg.Worker.DispatchTickInterval > 0 {
			workerOpts = append(workerOpts, worker.OptDispatchTickInterval(cfg.Worker.DispatchTickInterval))
		}
		if cfg.Worker.FlushInterval > 0 {
			workerOpts = append(workerOpts, worker.OptFlushInterval(cfg.Worker.FlushInterval))
		}
		w := worker.New(cfg.Hostname, modelMgr, workerOpts...)
		if cfg.ExpvarListenAddr != "" {
			w.Vars().Publish()
			go func() {
				if err := http.ListenAndServe(cfg.ExpvarListenAddr, expvar.Handler()); err != nil {
					log.GetLogger(ctx).Error("expvar server error", log.Any("err", err))
				}
			}()
		}
		return w.Run(ctx)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
