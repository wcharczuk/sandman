package server

import (
	"context"
	"sandman/pkg/model"

	sandmanv1 "sandman/proto/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WorkerServer struct {
	sandmanv1.WorkersServer
	Model *model.Manager
}

func (s WorkerServer) ListWorkers(ctx context.Context, args *sandmanv1.ListWorkersArgs) (*sandmanv1.ListWorkersResponse, error) {
	workers, err := s.Model.GetWorkers(ctx, args.LastSeenAfter.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	output := sandmanv1.ListWorkersResponse{
		Workers: make([]*sandmanv1.Worker, 0, len(workers)),
	}
	for _, t := range workers {
		output.Workers = append(output.Workers, s.protoWorkerFromModel(t))
	}
	return &output, nil
}

func (s WorkerServer) protoWorkerFromModel(t model.Worker) *sandmanv1.Worker {
	output := &sandmanv1.Worker{
		Hostname:    t.Hostname,
		CreatedUtc:  timestamppb.New(t.CreatedUTC),
		LastSeenUtc: timestamppb.New(t.LastSeenUTC),
	}
	return output
}
