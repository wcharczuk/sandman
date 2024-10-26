package main

import (
	"context"
	"expvar"
	"net/http"
	"os"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/db/migration"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"

	"sandman/pkg/config"
	"sandman/pkg/model"
	"sandman/pkg/scheduler"
)

func expvarBindAddr() string {
	if addr := os.Getenv("EXPVAR_BIND_ADDR"); addr != "" {
		return addr
	}
	return ":8080"
}

var entrypoint = apputil.DBEntryPoint[config.Config]{
	Setup: func(ctx context.Context, cfg config.Config) error {
		return dbutil.CreateDatabaseIfNotExists(ctx, cfg.DB.Database, db.OptLog(log.GetLogger(ctx)))
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return model.Migrations(migration.OptLog(log.GetLogger(ctx))).Apply(ctx, dbc)
	},
	Start: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-scheduler")
		modelMgr := &model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}
		s := scheduler.New(cfg.Hostname, modelMgr)
		s.Vars().Publish()
		go func() {
			if err := http.ListenAndServe(expvarBindAddr(), expvar.Handler()); err != nil {
				log.GetLogger(ctx).Error("expvar server error", log.Any("err", err))
			}
		}()
		return s.Run(ctx)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
