package worker

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.charczuk.com/sdk/async"
	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

// New returns a new worker.
func New(identity string, mgr *model.Manager) *Worker {
	return &Worker{
		identity: identity,
		mgr:      mgr,
		http:     new(http.Transport),
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

	http *http.Transport

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
	timers, err := w.mgr.GetDueTimers(ctx, w.identity)
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

		started := time.Now()
		var res *http.Response
		res, remoteErr = w.makeHookRequest(t)
		if remoteErr != nil || res.StatusCode >= http.StatusBadRequest {
			var statusCode int
			if res != nil {
				statusCode = res.StatusCode
			}
			if remoteErr != nil {
				log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to deliver to remote: %w", remoteErr), w.logAttrs(t,
					log.String("err_type", "remote"),
					log.Duration("elapsed", time.Since(started)),
				)...)
			} else {
				log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to deliver to remote: non-200 status code returned %d", statusCode), w.logAttrs(t,
					log.String("err_type", "remote"),
					log.Duration("elapsed", time.Since(started)),
				)...)
			}
			internalErr = w.mgr.MarkAttempted(ctx, t.ID, uint32(statusCode), remoteErr)
			if internalErr != nil {
				log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to mark attempted: %w", internalErr), w.logAttrs(t,
					log.String("err_type", "internal"),
					log.Duration("elapsed", time.Since(started)),
				)...)
			}
			return nil
		}
		internalErr = w.mgr.MarkDelivered(ctx, t.ID)
		if internalErr != nil {
			log.GetLogger(ctx).Err(fmt.Errorf("worker; failed to mark delivered: %w", internalErr), w.logAttrs(t,
				log.String("err_type", "internal"),
				log.Duration("elapsed", time.Since(started)),
			)...)
		}
		log.GetLogger(ctx).Info("worker; delivery success", w.logAttrs(t,
			log.Duration("elapsed", time.Since(started)),
		)...)
		return nil
	}
}

func (w *Worker) makeHookRequest(t *model.Timer) (*http.Response, error) {
	var body io.Reader
	if rawBody := t.HookBody; len(rawBody) > 0 {
		body = bytes.NewReader(rawBody)
	}

	var method string
	if hookMethod := t.HookMethod; hookMethod != "" {
		method = hookMethod
	} else {
		method = http.MethodPost
	}

	requestContext, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	req, err := http.NewRequestWithContext(requestContext, method, t.HookURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hook details: %w", err)
	}
	req.Header = w.metadata(t)
	client := &http.Client{
		Transport: w.http,
	}
	return client.Do(req)
}

func (w *Worker) logAttrs(t *model.Timer, extra ...any) []any {
	return append([]any{
		log.String("id", t.ID.String()),
		log.String("name", t.Name),
		log.String("method", t.HookMethod),
		log.String("url", t.HookURL),
	}, extra...)
}

func (w *Worker) metadata(t *model.Timer) (output http.Header) {
	output = make(http.Header)
	for key, value := range t.HookHeaders {
		output.Set(key, value)
	}
	return
}
