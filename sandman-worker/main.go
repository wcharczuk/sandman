package main

import (
	"context"
	"expvar"
	"net/http"
	"os"
	"sandman/pkg/config"
	"sandman/pkg/model"
	"sandman/pkg/worker"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/db/migration"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"
)

var entrypoint = apputil.DBEntryPoint[config.Config]{
	Setup: func(ctx context.Context, cfg config.Config) error {
		return dbutil.CreateDatabaseIfNotExists(ctx, cfg.DB.Database, db.OptLog(log.GetLogger(ctx)))
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return model.Migrations(migration.OptLog(log.GetLogger(ctx))).Apply(ctx, dbc)
	},
	Start: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-worker")
		modelMgr := model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}
		w := worker.NewWorker(0, "worker", modelMgr)
		w.Vars().Publish()
		go func() {
			if err := http.ListenAndServe(":8080", expvar.Handler()); err != nil {
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
