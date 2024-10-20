package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strings"

	"go.charczuk.com/sdk/cliutil"
	"go.charczuk.com/sdk/log"
	"google.golang.org/grpc"

	protos "sandman/examples/echo-srv/proto"
	"sandman/pkg/grpcutil"
)

func bindAddr() string {
	if value := os.Getenv("BIND_ADDR"); value != "" {
		return value
	}
	return ":5555"
}

func main() {
	logger := log.New()
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcutil.Recover(),
			grpcutil.Logged(logger),
		),
	)
	protos.RegisterEchoServer(s, new(echoServer))

	addr := bindAddr()
	var socketListener net.Listener
	var err error
	if strings.HasPrefix(addr, "unix://") {
		socketListener, err = net.Listen("unix", strings.TrimPrefix(addr, "unix://"))
	} else {
		socketListener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		cliutil.Fatal(err)
	}
	slog.Info("listening", "addr", addr)
	if err := s.Serve(socketListener); err != nil {
		cliutil.Fatal(err)
	}
}

type echoServer struct {
	protos.EchoServer
}

func Greet(ctx context.Context, args *protos.GreetArgs) (*protos.GreetResponse, error) {
	return &protos.GreetResponse{
		Response: "hello " + args.Name + "!",
	}, nil
}
