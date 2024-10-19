package server

import (
	"context"
	"sandman/pkg/model"
	sandmanv1 "sandman/proto/v1"
	"time"

	"go.charczuk.com/sdk/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TimerServer struct {
	sandmanv1.TimersServer
	Model model.Manager
}

func (s TimerServer) CreateTimer(ctx context.Context, t *sandmanv1.Timer) (*emptypb.Empty, error) {
	newTimer := model.Timer{
		ID:           uuid.V4(),
		Name:         t.GetName(),
		ShardID:      t.GetShardId(),
		Labels:       t.GetLabels(),
		CreatedUTC:   time.Now().UTC(),
		DueUTC:       t.GetDueUtc().AsTime(),
		RPCAddr:      t.GetRpcAddr(),
		RPCAuthority: t.GetRpcAuthority(),
		RPCMethod:    t.GetRpcMethod(),
		RPCMeta:      t.GetRpcMeta(),
		RPCArgs:      t.GetRpcArgs(),
	}
	if err := s.Model.CreateTimer(ctx, &newTimer); err != nil {
		err = status.Error(codes.Internal, err.Error())
		return nil, err
	}
	return nil, nil
}

func (s TimerServer) ListTimers(context.Context, *sandmanv1.ListTimersArgs) (*sandmanv1.ListTimersResponse, error) {
	return nil, nil
}

func (s TimerServer) GetTimer(context.Context, *sandmanv1.GetTimerArgs) (*sandmanv1.Timer, error) {
	return nil, nil
}

func (s TimerServer) DeleteTimer(context.Context, *sandmanv1.DeleteTimerArgs) (*emptypb.Empty, error) {
	return nil, nil
}
