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
		return nil
	},
	Migrate: func(ctx context.Context, cfg config.Config, dbc *db.Connection) error {
		return nil
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
		var interceptors = []grpc.UnaryServerInterceptor{
			grpcutil.Recover(),
			grpcutil.Logged(logger),
		}

		serverOpts := grpc.ChainUnaryInterceptor(interceptors...)

		s := grpc.NewServer(
			serverOpts,
		)

		ts := server.TimerServer{Model: modelMgr}
		v1.RegisterTimersServer(s, ts)
		ws := server.WorkerServer{Model: modelMgr}
		v1.RegisterWorkersServer(s, ws)

		bindAddr := cfg.Server.BindAddr
		var socketListener net.Listener
		var err error
		if after, ok := strings.CutPrefix(bindAddr, "unix://"); ok {
			socketListener, err = net.Listen("unix", after)
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
