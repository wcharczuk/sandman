package worker

import (
	"bytes"
	"context"
	"errors"
	"expvar"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.charczuk.com/sdk/async"
	"go.charczuk.com/sdk/log"
	"go.charczuk.com/sdk/uuid"

	"sandman/pkg/model"
	"sandman/pkg/utils"
	"sandman/pkg/wheel"
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

func OptPollingInterval(tickInterval time.Duration) WorkerOption {
	return func(w *Worker) {
		w.pollingInterval = tickInterval
	}
}

func OptHookTimeout(timeout time.Duration) WorkerOption {
	return func(w *Worker) {
		w.hookTimeout = timeout
	}
}

func OptBatchSize(batchSize int) WorkerOption {
	return func(w *Worker) {
		w.batchSize = batchSize
	}
}

// OptPrefetchWindow turns on wheel-mode dispatch. Each prefetch claims
// every timer due within the next `window` and parks them in an in-
// memory hash wheel; a 1-second dispatch loop drains the wheel and a
// flush loop batches DB writes. Zero (the default) keeps the legacy
// single-loop "claim due now, fire, mark delivered" behavior.
//
// The prefetch query widens its lease to window+leaseSafetyMargin so
// the worker can hold a timer in the wheel for the full window without
// the row aging back into the reclaim path. Setting window larger than
// the wheel's slot count is a configuration error and will be clamped.
func OptPrefetchWindow(window time.Duration) WorkerOption {
	return func(w *Worker) {
		w.prefetchWindow = window
	}
}

// OptDispatchTickInterval overrides the wheel-mode dispatch cadence.
// Defaults to 1s; only meaningful when OptPrefetchWindow > 0.
func OptDispatchTickInterval(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.dispatchTickInterval = d
	}
}

// OptFlushInterval overrides how often the wheel-mode flush loop pushes
// batched delivered/attempted writes to the DB. Defaults to 1s.
func OptFlushInterval(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.flushInterval = d
	}
}

type Worker struct {
	identity string
	mgr      *model.Manager

	parallelism     int
	pollingInterval time.Duration
	hookTimeout     time.Duration
	batchSize       int

	prefetchWindow       time.Duration
	dispatchTickInterval time.Duration
	flushInterval        time.Duration

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
	defer w.deregister(ctx)
	if w.prefetchWindow > 0 {
		return w.runWheelMode(ctx)
	}
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

// deregister removes this worker's row from the workers table on
// graceful shutdown so peers can reclaim its shard band immediately
// instead of waiting for last_seen_utc to age out. Uses a fresh
// context.Background() with a short timeout because the Run ctx is
// already cancelled by the time this runs; logger is retrieved from
// the (now-cancelled) parent where it was originally attached.
func (w *Worker) deregister(parent context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.mgr.DeleteWorker(ctx, w.identity); err != nil {
		log.GetLogger(parent).Error("worker; failed to deregister on shutdown", log.Any("err", err))
	}
}

const defaultParallelism = 255

func (w *Worker) parallelismOrDefault() int {
	if w.parallelism > 0 {
		return w.parallelism
	}
	return defaultParallelism
}

const defaultPollingInterval = 5 * time.Second

func (w *Worker) tickIntervalOrDefault() time.Duration {
	if w.pollingInterval > 0 {
		return w.pollingInterval
	}
	return defaultPollingInterval
}

const defaultHookTimeout = 1 * time.Second

func (w *Worker) hookTimeoutOrDefault() time.Duration {
	if w.hookTimeout > 0 {
		return w.hookTimeout
	}
	return defaultHookTimeout
}

const defaultBatchSize = 255

func (w *Worker) batchSizeOrDefault() int {
	if w.batchSize > 0 {
		return w.batchSize
	}
	return defaultBatchSize
}

func (w *Worker) processTick(ctx context.Context) {
	nowUTC := time.Now().UTC()

	if err := w.mgr.WorkerSeen(ctx, w.identity, nowUTC); err != nil {
		log.GetLogger(ctx).Error("worker; failed to update last seen", log.Any("err", err))
		return
	}

	shardLo, shardHi := w.currentShardBand(ctx, nowUTC)

	timers, err := w.mgr.GetDueTimers(ctx, w.identity, nowUTC, w.batchSizeOrDefault(), shardLo, shardHi)
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
		w.bulkMarkDeliveredWithRetry(ctx, deliveredIDs)
	}
}

// shardStaleness is how recently a peer must have reported in to count
// as part of the partitioning ring. Generous relative to the polling
// interval so a transiently-slow worker doesn't drop out.
const shardStaleness = 30 * time.Second

// shardSpace is one past the max uint32 shard value; using uint64 so
// the upper bound is expressible without overflow.
const shardSpace uint64 = 1 << 32

// currentShardBand computes the half-open [lo, hi) slice of the uint32
// shard space this worker should poll this tick. Workers sort by
// hostname and each claims its indexed slice of the space. If the
// worker isn't yet in the visible set (first tick race, stale `workers`
// read) we fall back to the whole space — SKIP LOCKED keeps that safe.
func (w *Worker) currentShardBand(ctx context.Context, now time.Time) (uint64, uint64) {
	peers, err := w.mgr.GetWorkers(ctx, now.Add(-shardStaleness))
	if err != nil || len(peers) == 0 {
		if err != nil {
			log.GetLogger(ctx).Error("worker; failed to list peers for shard band", log.Any("err", err))
		}
		return 0, shardSpace
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].Hostname < peers[j].Hostname })
	myIndex := -1
	for i, p := range peers {
		if p.Hostname == w.identity {
			myIndex = i
			break
		}
	}
	if myIndex < 0 {
		return 0, shardSpace
	}
	n := uint64(len(peers))
	band := shardSpace / n
	lo := uint64(myIndex) * band
	hi := lo + band
	// Last worker absorbs the remainder so the whole space is covered.
	if myIndex == len(peers)-1 {
		hi = shardSpace
	}
	return lo, hi
}

const dbWriteMaxRetries = 3
const dbWriteTimeout = 10 * time.Second

// retriableSQLStates are SQLSTATE codes that indicate the statement
// should be re-run as-is: transient contention / serialization failures
// with no application-level cause to fix first.
var retriableSQLStates = map[string]struct{}{
	"40001": {}, // serialization_failure
	"40P01": {}, // deadlock_detected
	"55P03": {}, // lock_not_available
	"08006": {}, // connection_failure
	"08003": {}, // connection_does_not_exist
	"08001": {}, // sqlclient_unable_to_establish_sqlconnection
	"57P01": {}, // admin_shutdown
	"57P02": {}, // crash_shutdown
	"57P03": {}, // cannot_connect_now
}

// isRetriableDBError reports whether err was produced by a condition
// where re-running the same statement is the correct remedy: a known
// transient SQLSTATE, a pgx connection-level "safe to retry" signal, or
// our own per-attempt context deadline (the DB was slow, not wrong).
// Anything else surfaces immediately so we don't burn retries on logic
// or constraint errors.
func isRetriableDBError(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		_, ok := retriableSQLStates[pgErr.Code]
		return ok
	}
	if pgconn.SafeToRetry(err) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

// retryDBWrite runs fn up to dbWriteMaxRetries times, each with a fresh
// per-attempt deadline sourced from context.Background() so retries can
// outlive the caller's tick budget. Non-retriable errors are surfaced
// immediately instead of burning remaining attempts on them. Returns
// the final error (nil on success, otherwise the last attempt's err).
func retryDBWrite(ctx context.Context, label string, fn func(context.Context) error) error {
	logger := log.GetLogger(ctx)
	var lastErr error
	for attempt := range dbWriteMaxRetries {
		writeCtx, cancel := context.WithTimeout(context.Background(), dbWriteTimeout)
		lastErr = fn(writeCtx)
		cancel()
		if lastErr == nil {
			return nil
		}
		retriable := isRetriableDBError(lastErr)
		logger.Error(label,
			log.Int("attempt", attempt+1),
			log.Int("max_attempts", dbWriteMaxRetries),
			log.Bool("retriable", retriable),
			log.Any("err", lastErr),
		)
		if !retriable {
			return lastErr
		}
	}
	return lastErr
}

func (w *Worker) bulkMarkDeliveredWithRetry(ctx context.Context, ids []uuid.UUID) error {
	return retryDBWrite(ctx, "worker; failed to mark timers delivered", func(c context.Context) error {
		return w.mgr.BulkMarkDelivered(c, time.Now().UTC(), ids)
	})
}

func (w *Worker) markAttemptedWithRetry(ctx context.Context, id uuid.UUID, statusCode uint32, remoteErr error, asOf time.Time) error {
	return retryDBWrite(ctx, "worker; failed to mark attempted", func(c context.Context) error {
		return w.mgr.MarkAttempted(c, id, statusCode, remoteErr, asOf)
	})
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
			internalErr = w.markAttemptedWithRetry(ctx, t.ID, uint32(statusCode), remoteErr, time.Now().UTC())
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

	requestContext, cancelTimeout := context.WithTimeout(context.Background(), w.hookTimeoutOrDefault())
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

//
// wheel-mode dispatch
//

// wheelLeaseSafetyMargin is added to the prefetch window to size the
// per-claim lease. The wheel must be allowed to hold a timer for the
// full window plus dispatch + retry budget; if the worker dies, the
// reclaim path picks the row back up after this many seconds elapse
// past the original "due-by" cutoff.
const wheelLeaseSafetyMargin = 30 * time.Second

// wheelSlotHeadroom is how many extra one-second slots the wheel keeps
// on top of the prefetch window. Inserts straddling a tick boundary
// land in slots ahead of the cursor; the headroom guarantees we can't
// hash an inserted timer into the slot the cursor just emptied.
const wheelSlotHeadroom = 4

const defaultDispatchTickInterval = 1 * time.Second

func (w *Worker) dispatchTickIntervalOrDefault() time.Duration {
	if w.dispatchTickInterval > 0 {
		return w.dispatchTickInterval
	}
	return defaultDispatchTickInterval
}

const defaultFlushInterval = 1 * time.Second

func (w *Worker) flushIntervalOrDefault() time.Duration {
	if w.flushInterval > 0 {
		return w.flushInterval
	}
	return defaultFlushInterval
}

// dispatchResult is the outcome of a single hook firing, queued onto the
// flushLoop's results channel. Either DeliveredAt is set (success) or
// StatusCode/RemoteErr describe the failure to record.
type dispatchResult struct {
	ID          uuid.UUID
	DeliveredAt time.Time
	StatusCode  uint32
	RemoteErr   error
	Failed      bool
}

func (w *Worker) runWheelMode(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	now := time.Now().UTC()
	slotCount := max(int(w.prefetchWindow/time.Second)+wheelSlotHeadroom, 8)
	wh := wheel.New(slotCount, now)
	results := make(chan dispatchResult, 4096)

	logger.Info("worker; wheel mode start",
		log.Duration("prefetch_window", w.prefetchWindow),
		log.Int("slot_count", slotCount),
		log.Duration("dispatch_tick", w.dispatchTickIntervalOrDefault()),
		log.Duration("flush_interval", w.flushIntervalOrDefault()),
	)

	var loops sync.WaitGroup
	loops.Add(3)
	go func() { defer loops.Done(); w.prefetchLoop(ctx, wh) }()
	go func() {
		defer loops.Done()
		// dispatch is the sole producer; closing results when it
		// returns lets flushLoop drain in-flight sends and exit
		// cleanly without losing the last batch's outcomes.
		defer close(results)
		w.dispatchLoop(ctx, wh, results)
	}()
	go func() { defer loops.Done(); w.flushLoop(ctx, results) }()
	loops.Wait()

	w.shutdownWheel(wh)
	return nil
}

// prefetchLoop is the only loop that talks to the DB on the read path.
// Each tick claims everything due within the prefetch window and parks
// it in the wheel; a follow-up prefetch deduplicates on timer ID so an
// overlapping window can't double-fire a row.
func (w *Worker) prefetchLoop(ctx context.Context, wh *wheel.Wheel) {
	tick := time.NewTicker(w.tickIntervalOrDefault())
	defer tick.Stop()
	logger := log.GetLogger(ctx)

	prefetch := func() {
		nowUTC := time.Now().UTC()
		if err := w.mgr.WorkerSeen(ctx, w.identity, nowUTC); err != nil {
			logger.Error("worker; failed to update last seen", log.Any("err", err))
			return
		}
		shardLo, shardHi := w.currentShardBand(ctx, nowUTC)
		windowSeconds := int(w.prefetchWindow / time.Second)
		leaseSeconds := windowSeconds + int(wheelLeaseSafetyMargin/time.Second)
		// Scale the claim batch so we hit the same wall-clock throughput
		// regardless of how wide the window is. Without the scale the
		// batch caps at the same count we previously claimed every 5s,
		// starving the wheel when the window is large.
		batch := w.batchSizeOrDefault()
		if windowSeconds > int(w.tickIntervalOrDefault()/time.Second) {
			scale := windowSeconds / int(w.tickIntervalOrDefault()/time.Second)
			if scale > 1 {
				batch = batch * scale
			}
		}

		timers, err := w.mgr.GetDueTimersWindowed(ctx, w.identity, nowUTC, batch, shardLo, shardHi, windowSeconds, leaseSeconds)
		if err != nil {
			logger.Error("worker; failed to prefetch timers", log.Any("err", err))
			return
		}
		var inserted, dropped int
		for i := range timers {
			if wh.Insert(&timers[i]) {
				inserted++
			} else {
				dropped++
			}
		}
		if inserted > 0 || dropped > 0 {
			logger.Info("worker; wheel prefetch",
				log.Int("inserted", inserted),
				log.Int("dropped", dropped),
				log.Int("wheel_len", wh.Len()),
			)
		}
	}

	prefetch() // run once immediately so the wheel isn't empty for the first dispatch tick.
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			prefetch()
		}
	}
}

// dispatchLoop drives the wheel cursor and fires hooks for every timer
// whose slot has come due. Concurrency is bounded by parallelism so a
// large slot can't fan out beyond the configured limit; results stream
// onto a channel for the flush loop to batch.
func (w *Worker) dispatchLoop(ctx context.Context, wh *wheel.Wheel, results chan<- dispatchResult) {
	tick := time.NewTicker(w.dispatchTickIntervalOrDefault())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			fired := wh.Advance(time.Now().UTC())
			if len(fired) == 0 {
				continue
			}
			b, _ := async.BatchContext(ctx)
			b.SetLimit(w.parallelismOrDefault())
			for i := range fired {
				t := fired[i]
				b.Go(func() error {
					w.fireOne(ctx, t, results)
					return nil
				})
			}
			_ = b.Wait()
		}
	}
}

// fireOne issues the hook request for t and pushes the outcome onto
// results. The flush loop is responsible for the corresponding DB
// write; failures and successes alike must always reach the channel
// so a stuck flush can't silently lose attempts.
func (w *Worker) fireOne(ctx context.Context, t *model.Timer, results chan<- dispatchResult) {
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
	res, remoteErr := w.makeHookRequest(t)
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
		select {
		case results <- dispatchResult{ID: t.ID, Failed: true, StatusCode: uint32(statusCode), RemoteErr: remoteErr}:
		case <-ctx.Done():
		}
		return
	}
	select {
	case results <- dispatchResult{ID: t.ID, DeliveredAt: time.Now().UTC()}:
	case <-ctx.Done():
	}
}

// flushLoop coalesces dispatch results into batched DB writes. Success
// rows go through BulkMarkDelivered; failures group by (status, err)
// so each "family" of failure costs one UPDATE per flush instead of
// one per timer. Termination is keyed off the results channel closing
// (dispatch loop is the sole producer) rather than ctx.Done — that
// guarantees the very last batch's outcomes always reach the DB even
// if shutdown races with a tick boundary.
func (w *Worker) flushLoop(ctx context.Context, results <-chan dispatchResult) {
	tick := time.NewTicker(w.flushIntervalOrDefault())
	defer tick.Stop()

	var delivered []uuid.UUID
	type failureKey struct {
		status uint32
		errMsg string
	}
	failures := map[failureKey][]uuid.UUID{}
	logger := log.GetLogger(ctx)

	flush := func() {
		if len(delivered) > 0 {
			w.bulkMarkDeliveredWithRetry(ctx, delivered)
			logger.Info("worker; flushed deliveries", log.Int("count", len(delivered)))
			delivered = delivered[:0]
		}
		for k, ids := range failures {
			if len(ids) == 0 {
				continue
			}
			err := retryDBWrite(ctx, "worker; failed to mark batch attempted", func(c context.Context) error {
				return w.mgr.BulkMarkAttempted(c, k.status, errOrNil(k.errMsg), time.Now().UTC(), ids)
			})
			if err == nil {
				logger.Info("worker; flushed attempts",
					log.Int("count", len(ids)),
					log.Int("status", int(k.status)),
				)
			}
			delete(failures, k)
		}
	}

	for {
		select {
		case r, ok := <-results:
			if !ok {
				flush()
				return
			}
			if r.Failed {
				k := failureKey{status: r.StatusCode, errMsg: errString(r.RemoteErr)}
				failures[k] = append(failures[k], r.ID)
			} else {
				delivered = append(delivered, r.ID)
			}
		case <-tick.C:
			flush()
		}
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// errOrNil reverses errString — used when handing the grouped-failure
// message back to BulkMarkAttempted, which itself stringifies. Wrapping
// in a synthetic error keeps the manager's signature unchanged.
func errOrNil(s string) error {
	if s == "" {
		return nil
	}
	return errString_synthetic{msg: s}
}

type errString_synthetic struct{ msg string }

func (e errString_synthetic) Error() string { return e.msg }

// shutdownWheel is invoked after Run's loops exit. It tries, on a fresh
// context, to relinquish the wheel's still-claimed contents so peers
// can pick them up immediately on their next prefetch instead of
// waiting the full lease out.
func (w *Worker) shutdownWheel(wh *wheel.Wheel) {
	left := wh.DrainAll()
	if len(left) == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, len(left))
	for _, t := range left {
		ids = append(ids, t.ID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.mgr.BulkRelinquish(ctx, w.identity, time.Now().UTC(), ids); err != nil {
		log.GetLogger(ctx).Error("worker; failed to relinquish wheel on shutdown",
			log.Int("count", len(ids)),
			log.Any("err", err),
		)
	}
}
