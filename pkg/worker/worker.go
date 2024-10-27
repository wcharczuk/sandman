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
	"go.charczuk.com/sdk/uuid"

	"sandman/pkg/model"
	"sandman/pkg/utils"
)

// New returns a new worker.
func New(identity string, mgr *model.Manager, opts ...WorkerOption) *Worker {
	w := &Worker{
		identity: identity,
		mgr:      mgr,
		http:     new(http.Transport),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
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

func OptBatchSize(batchSize int) WorkerOption {
	return func(w *Worker) {
		w.batchSize = batchSize
	}
}

type Worker struct {
	identity string
	mgr      *model.Manager

	parallelism  int
	tickInterval time.Duration
	batchSize    int

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

const defaultBatchSize = 255

func (w *Worker) batchSizeOrDefault() int {
	if w.batchSize > 0 {
		return w.batchSize
	}
	return defaultBatchSize
}

func (w *Worker) processTick(ctx context.Context) {
	timers, err := w.mgr.GetDueTimers(ctx, w.identity, w.batchSizeOrDefault())
	if err != nil {
		log.GetLogger(ctx).Error("worker; failed to get timers", log.Any("err", err))
		return
	}
	b, _ := async.BatchContext(ctx)
	b.SetLimit(w.parallelismOrDefault())
	for index := range timers {
		b.Go(w.processTickTimer(ctx, &timers[index]))
	}

	if err := b.Wait(); err != nil {
		log.GetLogger(ctx).Error("worker; failed to process timers", log.Any("err", err))
		return
	}

	var deliveredIDs []uuid.UUID
	for index := range timers {
		if timers[index].DeliveredUTC != nil && !timers[index].DeliveredUTC.IsZero() {
			deliveredIDs = append(deliveredIDs, timers[index].ID)
		}
	}

	if len(deliveredIDs) > 0 {
		log.GetLogger(ctx).Info("worker; marking timers delivered",
			log.Int("timers", len(deliveredIDs)),
		)
		if err := w.mgr.BulkUpdateTimerSuccesses(ctx, time.Now().UTC(), deliveredIDs); err != nil {
			log.GetLogger(ctx).Error("worker; failed to mark timers delivered", log.Any("err", err))
		}
	}
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

		// mark the timer as delivered
		t.DeliveredUTC = utils.Ref(time.Now().UTC())
		return nil
	}
}

// makeHookRequest makes the underlying request to the hook url, with a 5 second timeout.
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
