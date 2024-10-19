package grpcutil

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Service struct {
	listener net.Listener
	server   *grpc.Server
}

func (s Service) Start(ctx context.Context) error {
	return s.server.Serve(s.listener)
}

func (s Service) Restart(_ context.Context) error { return nil }

func (s Service) Stop(_ context.Context) error {
	s.server.GracefulStop()
	return nil
}
