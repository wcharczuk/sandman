package worker

import (
	"context"
	"expvar"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	"go.charczuk.com/sdk/async"
	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

// NewWorker returns a new worker.
func NewWorker(identity string, mgr *model.Manager) *Worker {
	return &Worker{
		identity: identity,
		mgr:      mgr,
		clients:  make(map[string]*grpc.ClientConn),
	}
}

type WorkerOption func(*Worker)

func OptParallelism(parallelism int) WorkerOption {
	return func(w *Worker) {
		w.parallelism = parallelism
	}
}

func OptTickInterval(tickInterval time.Duration) WorkerOption {
	return func(w *Worker) {
		w.tickInterval = tickInterval
	}
}

type Worker struct {
	identity string
	mgr      *model.Manager

	parallelism  int
	tickInterval time.Duration

	clientsMu sync.Mutex
	clients   map[string]*grpc.ClientConn

	timersProcessed              expvar.Int
	timersProcessedRemoteError   expvar.Int
	timersProcessedInternalError expvar.Int
}

type WorkerVars struct {
	TimersProcessed              *expvar.Int
	TimersProcessedRemoteError   *expvar.Int
	TimersProcessedInternalError *expvar.Int
}

func (wv WorkerVars) Publish() {
	expvar.Publish("timers_processed", wv.TimersProcessed)
	expvar.Publish("timers_processed_remote_error", wv.TimersProcessedRemoteError)
	expvar.Publish("timers_processed_internal_error", wv.TimersProcessedInternalError)
}

func (w *Worker) Vars() WorkerVars {
	return WorkerVars{
		TimersProcessed:              &w.timersProcessed,
		TimersProcessedRemoteError:   &w.timersProcessedRemoteError,
		TimersProcessedInternalError: &w.timersProcessedInternalError,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	tick := time.NewTicker(w.tickIntervalOrDefault())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			deadlineCtx, deadlineCancel := context.WithTimeout(ctx, w.tickIntervalOrDefault())
			go func() {
				defer deadlineCancel()
				w.processTick(deadlineCtx)
			}()
		}
	}
}

const defaultTickInterval = 10 * time.Second

func (w *Worker) tickIntervalOrDefault() time.Duration {
	if w.tickInterval > 0 {
		return w.tickInterval
	}
	return defaultTickInterval
}

const defaultParallelism = 255

func (w *Worker) parallelismOrDefault() int {
	if w.parallelism > 0 {
		return w.parallelism
	}
	return defaultParallelism
}

func (w *Worker) processTick(ctx context.Context) {
	timers, err := w.mgr.GetDueTimers(ctx, w.identity, time.Now().UTC())
	if err != nil {
		log.GetLogger(ctx).Error("worker; failed to get timers", log.Any("err", err))
		return
	}
	b, _ := async.BatchContext(ctx)
	b.SetLimit(w.parallelismOrDefault())
	for index := range timers {
		b.Go(w.processTickTimer(ctx, &timers[index]))
	}
	b.Wait()
}

func (w *Worker) processTickTimer(ctx context.Context, t *model.Timer) func() error {
	return func() error {
		var internalErr, remoteErr error
		defer func() {
			w.timersProcessed.Add(1)
			if remoteErr != nil {
				w.timersProcessedRemoteError.Add(1)
			}
			if internalErr != nil {
				w.timersProcessedInternalError.Add(1)
			}
		}()
		var c *grpc.ClientConn
		c, remoteErr = w.clientForAddr(t.RPCAddr, t.RPCAuthority)
		if remoteErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to create client: %w", remoteErr), w.logAttrs(t, log.String("err_type", "remote"))...)
			return nil
		}
		ctx = metadata.NewOutgoingContext(ctx, w.metadata(t))

		req, internalErr := w.getArgs(t.RPCArgsTypeURL, t.RPCArgsData)
		if internalErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to create rpc args: %w", internalErr), w.logAttrs(t, log.String("err_type", "internal"))...)
			return nil
		}

		res, internalErr := w.getReturn(t.RPCReturnTypeURL)
		if internalErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to create rpc return: %w", internalErr), w.logAttrs(t, log.String("err_type", "internal"))...)
			return nil
		}

		remoteErr = c.Invoke(ctx, t.RPCMethod, req, &res, grpc.ForceCodec(new(rawCodec)))
		if remoteErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to deliver to remote: %w", remoteErr), w.logAttrs(t, log.String("err_type", "remote"))...)
			deliveredStatus, deliveredErr := w.formatDeliveredStatus(remoteErr)
			internalErr = w.mgr.MarkAttempted(ctx, t.ID, deliveredStatus, deliveredErr)
			if internalErr != nil {
				log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to mark attempted: %w", internalErr), w.logAttrs(t, log.String("err_type", "internal"))...)
			}
			return nil
		}
		internalErr = w.mgr.MarkDelivered(ctx, t.ID)
		if internalErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to mark delivered: %w", internalErr), w.logAttrs(t, log.String("err_type", "internal"))...)
		}

		log.GetLogger(ctx).Info("worker; delivery success", w.logAttrs(t)...)
		return nil
	}
}

func (w *Worker) getArgs(typeURL string, data []byte) (dst proto.Message, err error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByURL(typeURL)
	if err != nil {
		if err == protoregistry.NotFound {
			return nil, err
		}
		return nil, fmt.Errorf("could not resolve %q: %v", typeURL, err)
	}
	dst = mt.New().Interface()
	err = proto.Unmarshal(data, dst)
	return
}

func (w *Worker) getReturn(typeURL string) (dst proto.Message, err error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByURL(typeURL)
	if err != nil {
		if err == protoregistry.NotFound {
			return nil, err
		}
		return nil, fmt.Errorf("could not resolve %q: %v", typeURL, err)
	}
	dst = mt.New().Interface()
	return
}

type rawCodec struct {
	encoding.Codec
}

func (rc rawCodec) Name() string { return "raw" }

func (rc rawCodec) Marshal(v any) ([]byte, error) {
	switch typed := v.(type) {
	case []byte:
		return typed, nil
	case string:
		return []byte(typed), nil
	}
	return nil, fmt.Errorf("invalid raw type: %T", v)
}

func (rc rawCodec) Unmarshal(data []byte, v any) error {
	switch typed := v.(type) {
	case []byte:
		copy(typed, data)
		return nil
	case *[]byte:
		copy(*typed, data)
		return nil
	}
	return fmt.Errorf("invalid raw type: %T", v)
}

func (w *Worker) logAttrs(t *model.Timer, extra ...any) []any {
	return append([]any{
		log.String("id", t.ID.String()),
		log.String("name", t.Name),
		log.String("addr", t.RPCAddr),
		log.String("authority", t.RPCAuthority),
		log.String("method", t.RPCMethod),
	}, extra...)
}

func (w *Worker) metadata(t *model.Timer) (output metadata.MD) {
	output = make(metadata.MD)
	for key, value := range t.RPCMeta {
		output.Set(key, value)
	}
	return
}

func (w *Worker) formatDeliveredStatus(err error) (outputStatus uint32, outputErr string) {
	outputErr = fmt.Sprintf("%+v", err)
	s, ok := status.FromError(err)
	if !ok {
		return
	}
	outputStatus = uint32(s.Code())
	return
}

func (w *Worker) clientForAddr(addr, authority string) (*grpc.ClientConn, error) {
	if cached, ok := w.cachedClientForAddr(addr); ok {
		return cached, nil
	}
	newClient, err := grpc.NewClient(addr,
		grpc.WithAuthority(authority),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), // This sets the initial balancing policy.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	w.clientsMu.Lock()
	w.clients[addr] = newClient
	w.clientsMu.Unlock()
	return newClient, nil
}

func (w *Worker) cachedClientForAddr(addr string) (client *grpc.ClientConn, ok bool) {
	w.clientsMu.Lock()
	client, ok = w.clients[addr]
	w.clientsMu.Unlock()
	return
}
