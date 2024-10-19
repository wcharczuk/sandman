package main

import (
	"context"
	"net"
	"os"

	"go.charczuk.com/sdk/apputil"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/slant"
	"google.golang.org/grpc"

	"sandman/pkg/config"
	"sandman/pkg/model"
	"sandman/pkg/server"
	v1 "sandman/proto/v1"
)

var entrypoint = apputil.DBEntryPoint[config.Config]{
	Setup: func(ctx context.Context, cfg config.Config) error {
		return nil
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return nil
	},
	Start: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		slant.Print(os.Stdout, "sandman-srv")
		modelMgr := model.Manager{
			BaseManager: dbutil.NewBaseManager(dbc),
		}
		if err := modelMgr.Initialize(ctx); err != nil {
			return err
		}
		s := grpc.NewServer()
		ts := server.TimerServer{Model: modelMgr}
		v1.RegisterTimersServer(s, ts)
		l, err := net.Listen("tcp", ":8333")
		if err != nil {
			return err
		}
		return s.Serve(l)
	},
}

func init() {
	entrypoint.Init()
}

func main() {
	entrypoint.Main()
}
