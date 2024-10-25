package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.charczuk.com/sdk/selector"
	"go.charczuk.com/sdk/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sandman/pkg/model"
	sandmanv1 "sandman/proto/v1"
)

type TimerServer struct {
	sandmanv1.TimersServer
	Model *model.Manager
}

func (s TimerServer) CreateTimer(ctx context.Context, t *sandmanv1.Timer) (*sandmanv1.IdentifierResponse, error) {
	if t.GetDueUtc() == nil || t.GetDueUtc().AsTime().Before(time.Now().UTC()) {
		return nil, status.Error(codes.InvalidArgument, "invalid `due_utc`; must be set and in the future")
	}
	if t.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid `name`; must be set")
	}
	if t.GetHookUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid `hook_url`; must be set")
	}
	if _, err := url.Parse(t.GetHookUrl()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid `hook_url`; could not parse url")
	}
	if strings.ToLower(t.GetHookMethod()) == http.MethodGet && len(t.GetHookBody()) > 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid hook; `hook_method` cannot be GET with a body specified")
	}
	now := time.Now().UTC()
	dueUTC := t.GetDueUtc().AsTime()

	newTimer := model.Timer{
		Name:        t.GetName(),
		Labels:      t.GetLabels(),
		CreatedUTC:  now,
		DueUTC:      dueUTC,
		DueCounter:  minutesUntil(now, dueUTC),
		HookURL:     t.GetHookUrl(),
		HookMethod:  t.GetHookMethod(),
		HookHeaders: t.GetHookHeaders(),
		HookBody:    t.GetHookBody(),
	}
	if err := s.Model.Invoke(ctx).Create(&newTimer); err != nil {
		err = status.Error(codes.Internal, err.Error())
		return nil, err
	}
	return &sandmanv1.IdentifierResponse{
		Id: newTimer.ID.String(),
	}, nil
}

func minutesUntil(now, dueUTC time.Time) uint64 {
	diff := dueUTC.Sub(now)
	if diff <= 0 {
		return 0
	}
	return uint64(diff / time.Minute)
}

func (s TimerServer) ListTimers(ctx context.Context, args *sandmanv1.ListTimersArgs) (*sandmanv1.ListTimersResponse, error) {
	var compiledSelector selector.Selector
	var err error
	if rawSelector := args.GetSelector(); rawSelector != "" {
		compiledSelector, err = selector.Parse(rawSelector)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid selector; %v", err))
		}
	}
	before := time.Now().UTC().Add(time.Hour)
	after := time.Time{}

	if args.GetBefore() != nil && !args.GetBefore().AsTime().IsZero() {
		before = args.GetBefore().AsTime()
	}
	if args.GetAfter() != nil && !args.GetAfter().AsTime().IsZero() {
		after = args.GetAfter().AsTime()
	}

	timers, err := s.Model.GetTimersDueBetween(ctx, after, before, compiledSelector)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	output := sandmanv1.ListTimersResponse{
		Timers: make([]*sandmanv1.Timer, 0, len(timers)),
	}
	for _, t := range timers {
		output.Timers = append(output.Timers, s.protoTimerFromModel(t))
	}
	return &output, nil
}

func (s TimerServer) GetTimer(ctx context.Context, args *sandmanv1.GetTimerArgs) (*sandmanv1.Timer, error) {
	return s.getTimerByNameOrID(ctx, args.GetId(), args.GetName())
}

func (s TimerServer) DeleteTimer(ctx context.Context, args *sandmanv1.DeleteTimerArgs) (*emptypb.Empty, error) {
	var found bool
	var err error
	if id := args.GetId(); id != "" {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%q is not a valid uuid", id))
		}
		found, err = s.Model.DeleteTimerByID(ctx, parsedID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("timer with id %q not found", id))
		}
	} else if name := args.GetName(); name != "" {
		found, err = s.Model.DeleteTimerByName(ctx, name)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("timer with name %q not found", name))
		}
	}
	return nil, nil
}

//
// helpers
//

func (s TimerServer) getTimerByNameOrID(ctx context.Context, id, name string) (*sandmanv1.Timer, error) {
	if id != "" {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("%q is not a valid uuid", id))
		}
		var t model.Timer
		found, err := s.Model.Invoke(ctx).Get(&t, parsedID)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("timer with id %q not found", id))
		}
		return s.protoTimerFromModel(t), nil
	}
	if name != "" {
		t, found, err := s.Model.GetTimerByName(ctx, name)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("timer with name %q not found", name))
		}
		return s.protoTimerFromModel(t), nil
	}
	return nil, status.Error(codes.InvalidArgument, "one of `id` or `name` is required")
}

func (s TimerServer) protoTimerFromModel(t model.Timer) *sandmanv1.Timer {
	output := &sandmanv1.Timer{
		Id:                  t.ID.ShortString(),
		Name:                t.Name,
		Labels:              t.Labels,
		CreatedUtc:          timestamppb.New(t.CreatedUTC),
		DueUtc:              timestamppb.New(t.DueUTC),
		DueCounter:          t.DueCounter,
		RetryCounter:        t.RetryCounter,
		Attempt:             uint32(t.Attempt),
		HookUrl:             t.HookURL,
		HookMethod:          t.HookMethod,
		HookHeaders:         t.HookHeaders,
		HookBody:            t.HookBody,
		DeliveredStatusCode: t.DeliveredStatusCode,
		DeliveredErr:        t.DeliveredErr,
	}
	if t.AssignedWorker != nil && *t.AssignedWorker != "" {
		output.AssignedWorker = *t.AssignedWorker
	}
	if t.DeliveredUTC != nil && !t.DeliveredUTC.IsZero() {
		output.DeliveredUtc = timestamppb.New(*t.DeliveredUTC)
	}
	return output
}
