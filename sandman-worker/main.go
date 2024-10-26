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
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"
)

func expvarBindAddr() string {
	if addr := os.Getenv("EXPVAR_BIND_ADDR"); addr != "" {
		return addr
	}
	return ":8080"
}

var entrypoint = apputil.DBEntryPoint[config.Config]{
	Setup: func(ctx context.Context, cfg config.Config) error {
		return nil
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return nil
	},
	Start: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
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
			if err := http.ListenAndServe(expvarBindAddr(), expvar.Handler()); err != nil {
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
