package main

import (
	"context"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/db/migration"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/slant"

	"sandman/pkg/config"
	"sandman/pkg/grpcutil"
	"sandman/pkg/model"
	"sandman/pkg/server"
	v1 "sandman/proto/v1"
)

var entrypoint = apputil.DBEntryPoint[config.Config]{
	Setup: func(ctx context.Context, cfg config.Config) error {
		return dbutil.CreateDatabaseIfNotExists(ctx, cfg.DB.Database, db.OptLog(log.GetLogger(ctx)))
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return model.Migrations(migration.OptLog(log.GetLogger(ctx))).Apply(ctx, dbc)
	},
	Start: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-srv")
		modelMgr := &model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}
		logger := log.GetLogger(ctx)
		s := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpcutil.Recover(),
				grpcutil.Logged(logger),
			),
		)
		ts := server.TimerServer{Model: modelMgr}
		v1.RegisterTimersServer(s, ts)

		bindAddr := cfg.Server.BindAddr
		var socketListener net.Listener
		var err error
		if strings.HasPrefix(bindAddr, "unix://") {
			socketListener, err = net.Listen("unix", strings.TrimPrefix(bindAddr, "unix://"))
		} else {
			socketListener, err = net.Listen("tcp", bindAddr)
		}
		if err != nil {
			return err
		}
		logger.Info("listening", log.String("addr", bindAddr))
		return s.Serve(socketListener)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
