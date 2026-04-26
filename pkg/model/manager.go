package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/selector"
	"go.charczuk.com/sdk/uuid"
)

type Manager struct {
	dbutil.BaseManager

	getDueTimers        *sql.Stmt
	getTimerByName      *sql.Stmt
	getTimersDueBetween *sql.Stmt
	cullTimers          *sql.Stmt
	markAttempted       *sql.Stmt
	bulkMarkAttempted   *sql.Stmt
	bulkRelinquish      *sql.Stmt
	deleteTimerByID     *sql.Stmt
	deleteTimerByName   *sql.Stmt
	bulkMarkDelivered   *sql.Stmt
	workerSeen          *sql.Stmt
	deleteWorker          *sql.Stmt
	getWorkers            *sql.Stmt
	getPeakTimersDueCount *sql.Stmt
	getOverdueTimerCount  *sql.Stmt
}

func (m *Manager) Initialize(ctx context.Context) (err error) {
	m.getDueTimers, err = m.Invoke(ctx).Prepare(queryGetDueTimers)
	if err != nil {
		err = fmt.Errorf("getDueTimers: %w", err)
		return
	}
	m.getTimerByName, err = m.Invoke(ctx).Prepare(queryGetTimerByName)
	if err != nil {
		err = fmt.Errorf("getTimerByName: %w", err)
		return
	}
	m.getTimersDueBetween, err = m.Invoke(ctx).Prepare(queryGetTimersDueBetween)
	if err != nil {
		err = fmt.Errorf("getTimersDueBetween: %w", err)
		return
	}
	m.cullTimers, err = m.Invoke(ctx).Prepare(execCullTimers)
	if err != nil {
		err = fmt.Errorf("cullTimers: %w", err)
		return
	}
	m.markAttempted, err = m.Invoke(ctx).Prepare(execMarkAttempted)
	if err != nil {
		err = fmt.Errorf("markAttempted: %w", err)
		return
	}
	m.bulkMarkAttempted, err = m.Invoke(ctx).Prepare(execBulkMarkAttempted)
	if err != nil {
		err = fmt.Errorf("bulkMarkAttempted: %w", err)
		return
	}
	m.bulkRelinquish, err = m.Invoke(ctx).Prepare(execBulkRelinquish)
	if err != nil {
		err = fmt.Errorf("bulkRelinquish: %w", err)
		return
	}
	m.deleteTimerByID, err = m.Invoke(ctx).Prepare(execDeleteTimerByID)
	if err != nil {
		err = fmt.Errorf("deleteTimerByID: %w", err)
		return
	}
	m.deleteTimerByName, err = m.Invoke(ctx).Prepare(execDeleteTimerByName)
	if err != nil {
		err = fmt.Errorf("deleteTimerByName: %w", err)
		return
	}
	m.bulkMarkDelivered, err = m.Invoke(ctx).Prepare(execBulkMarkDelivered)
	if err != nil {
		err = fmt.Errorf("bulkMarkDelivered: %w", err)
		return
	}
	m.workerSeen, err = m.Invoke(ctx).Prepare(execWorkerSeen)
	if err != nil {
		err = fmt.Errorf("workerSeen: %w", err)
		return
	}
	m.deleteWorker, err = m.Invoke(ctx).Prepare(execDeleteWorker)
	if err != nil {
		err = fmt.Errorf("deleteWorker: %w", err)
		return
	}
	m.getWorkers, err = m.Invoke(ctx).Prepare(queryGetWorkers)
	if err != nil {
		err = fmt.Errorf("getWorkers: %w", err)
		return
	}
	m.getPeakTimersDueCount, err = m.Invoke(ctx).Prepare(queryGetPeakTimersDueCount)
	if err != nil {
		err = fmt.Errorf("getPeakTimersDueCount: %w", err)
		return
	}
	m.getOverdueTimerCount, err = m.Invoke(ctx).Prepare(queryGetOverdueTimerCount)
	if err != nil {
		err = fmt.Errorf("getOverdueTimerCount: %w", err)
		return
	}
	return
}

func (m Manager) Close() error {
	if err := m.cullTimers.Close(); err != nil {
		return err
	}
	if err := m.deleteTimerByID.Close(); err != nil {
		return err
	}
	if err := m.deleteTimerByName.Close(); err != nil {
		return err
	}
	if err := m.getDueTimers.Close(); err != nil {
		return err
	}
	if err := m.getTimerByName.Close(); err != nil {
		return err
	}
	if err := m.getTimersDueBetween.Close(); err != nil {
		return err
	}
	if err := m.markAttempted.Close(); err != nil {
		return err
	}
	if err := m.bulkMarkAttempted.Close(); err != nil {
		return err
	}
	if err := m.bulkRelinquish.Close(); err != nil {
		return err
	}
	if err := m.bulkMarkDelivered.Close(); err != nil {
		return err
	}
	if err := m.getPeakTimersDueCount.Close(); err != nil {
		return err
	}
	if err := m.getOverdueTimerCount.Close(); err != nil {
		return err
	}
	return nil
}

var queryGetTimerByName = fmt.Sprintf(`SELECT %s FROM %s WHERE name = $1`, db.ColumnNamesCSV(timerColumns), timerTableName)

func (m Manager) GetTimerByName(ctx context.Context, name string) (out Timer, found bool, err error) {
	rows, err := m.getTimerByName.QueryContext(ctx, name)
	if err != nil {
		err = rows.Err()
		return
	}
	if !rows.Next() {
		return
	}
	err = db.PopulateInOrder(&out, rows, timerColumns)
	return
}

var queryGetTimersDueBetween = fmt.Sprintf(`SELECT
	%s
FROM
	%s
WHERE
	due_utc > $1 AND due_utc < $2
`, db.ColumnNamesCSV(timerColumns), timerTableName)

func (m Manager) GetTimersDueBetween(ctx context.Context, after, before time.Time, s selector.Selector, opts ...db.InvocationOption) (output []Timer, err error) {
	var rows *sql.Rows
	rows, err = m.getTimersDueBetween.QueryContext(ctx, after, before)
	if err != nil {
		return
	}
	for rows.Next() {
		var t Timer
		if err = db.PopulateInOrder(&t, rows, timerColumns); err != nil {
			return
		}
		if s != nil {
			if s.Matches(t.MatchLabels()) {
				output = append(output, t)
			}
		} else {
			output = append(output, t)
		}
	}
	return
}

// queryGetDueTimers parameters:
//   $1 = worker identity, $2 = asOf, $3 = batch size,
//   $4 = shard low (inclusive), $5 = shard high (exclusive),
//   $6 = window seconds (>=0): widen the "due" cutoff to asOf + window
//        so callers can prefetch timers about to fire,
//   $7 = lease seconds: how long the claim should be held.
//
// The shuffle-shard priority boost still uses asOf alone so the
// per-tick fairness ordering is unchanged when window > 0.
var queryGetDueTimers = fmt.Sprintf(`WITH candidates AS (
	SELECT
		id, priority, shard, due_utc
	FROM
		%[1]s@ix_timers_shard_due_utc_pending
	WHERE
		shard >= $4 AND shard < $5
		AND due_utc < $2::timestamp + ($6 * interval '1 second')
		AND (
			assigned_until_utc IS NULL OR (assigned_until_utc IS NOT NULL AND assigned_until_utc < $2)
		)
		AND (
			retry_utc IS NULL OR (retry_utc IS NOT NULL AND retry_utc < $2)
		)
		AND attempt < 5
		AND delivered_utc IS NULL
	ORDER BY shard ASC, due_utc ASC
	LIMIT $3 * 2
	FOR UPDATE SKIP LOCKED
), selected AS (
	SELECT
		id
	FROM
		candidates
	ORDER BY
		priority
		+ (
			(
				MOD(shard, 3600) + cast(extract('minute', $2::timestamp) * extract('second', $2::timestamp) as BIGINT)
			)
			* 100
		)
		+ (cast(extract('epoch', $2::timestamp) - extract('epoch', due_utc) as BIGINT) * 100)
	DESC
	LIMIT $3
)
UPDATE %[1]s
SET
	assigned_worker = $1
	, attempt = attempt + 1
	, assigned_until_utc = $2::timestamp + ($7 * interval '1 second')
	, retry_utc = $2 + interval '5 minutes'
WHERE
	id in (SELECT id FROM selected)
RETURNING %[2]s
`, timerTableName, db.ColumnNamesCSV(timerColumns))

//

// defaultLeaseSeconds matches the original 1-minute claim window used by
// the legacy non-windowed path. Wheel-mode callers pass an explicit
// lease covering windowSeconds + safety margin via GetDueTimersWindowed.
const defaultLeaseSeconds = 60

// GetDueTimers claims up to batchSize timers whose shard falls in
// [shardLo, shardHi). The half-open range lets workers partition the
// uint32 shard space cleanly: a worker that owns the whole space passes
// (0, 1<<32). `FOR UPDATE SKIP LOCKED` still protects against two
// workers briefly overlapping bands during a membership change.
//
// This is the legacy "due now" entry point — equivalent to
// GetDueTimersWindowed with windowSeconds=0 and the default lease.
func (m Manager) GetDueTimers(ctx context.Context, workerIdentity string, asOf time.Time, batchSize int, shardLo, shardHi uint64) (output []Timer, err error) {
	return m.GetDueTimersWindowed(ctx, workerIdentity, asOf, batchSize, shardLo, shardHi, 0, defaultLeaseSeconds)
}

// GetDueTimersWindowed is the wheel-mode claim path. Callers can ask for
// timers due any time before asOf+windowSeconds and set the lease to
// cover the full window plus a safety margin (e.g. windowSeconds + 30).
// windowSeconds=0 reproduces GetDueTimers' "due now" behavior; the lease
// must always be positive so the partial-index reclaim path can still
// recover unfired timers if this worker dies.
func (m Manager) GetDueTimersWindowed(ctx context.Context, workerIdentity string, asOf time.Time, batchSize int, shardLo, shardHi uint64, windowSeconds, leaseSeconds int) (output []Timer, err error) {
	if leaseSeconds <= 0 {
		leaseSeconds = defaultLeaseSeconds
	}
	var rows *sql.Rows
	rows, err = m.getDueTimers.QueryContext(ctx, workerIdentity, asOf, batchSize, shardLo, shardHi, windowSeconds, leaseSeconds)
	if err != nil {
		return
	}
	for rows.Next() {
		var t Timer
		if err = db.PopulateInOrder(&t, rows, timerColumns); err != nil {
			return
		}
		output = append(output, t)
	}
	return
}

var execCullTimers = fmt.Sprintf(`DELETE FROM %s 
WHERE 
	(delivered_utc IS NOT NULL OR attempt >= 5)
	AND due_utc < $1
`, timerTableName)

func (m Manager) CullTimers(ctx context.Context, cutoff time.Time) (rowsAffected int64, err error) {
	res, err := m.cullTimers.ExecContext(ctx, cutoff)
	if err != nil {
		return
	}
	rowsAffected, _ = res.RowsAffected()
	return
}

var execMarkAttempted = fmt.Sprintf(`UPDATE %s
SET
	delivered_status_code = $2
	, delivered_err = $3
	, retry_utc = $4::timestamp + interval '5 minutes'
	, assigned_until_utc = NULL
WHERE
	id = $1
	AND attempt < 5
`, timerTableName)

func (m Manager) MarkAttempted(ctx context.Context, id uuid.UUID, deliveredStatus uint32, deliveredErr error, asOf time.Time) (err error) {
	var deliveredErrString string
	if deliveredErr != nil {
		deliveredErrString = deliveredErr.Error()
	}
	_, err = m.markAttempted.ExecContext(ctx, id, deliveredStatus, deliveredErrString, asOf)
	return
}

// execBulkMarkAttempted records identical (status, err) for a batch of
// timer IDs. The dispatcher coalesces failures by (status_code, err_msg)
// so the wheel-mode flush loop can write one row per failure family
// instead of one per timer; the per-timer execMarkAttempted above is
// preserved for the legacy single-shot tick path.
var execBulkMarkAttempted = fmt.Sprintf(`UPDATE %s
SET
	delivered_status_code = $1
	, delivered_err = $2
	, retry_utc = $3::timestamp + interval '5 minutes'
	, assigned_until_utc = NULL
WHERE
	id = ANY($4)
	AND attempt < 5
`, timerTableName)

func (m Manager) BulkMarkAttempted(ctx context.Context, deliveredStatus uint32, deliveredErr error, asOf time.Time, ids []uuid.UUID) (err error) {
	if len(ids) == 0 {
		return
	}
	var deliveredErrString string
	if deliveredErr != nil {
		deliveredErrString = deliveredErr.Error()
	}
	_, err = m.bulkMarkAttempted.ExecContext(ctx, deliveredStatus, deliveredErrString, asOf, ids)
	return
}

// execBulkRelinquish drops the worker's claim on a batch of timers
// without recording an attempt. Used on graceful shutdown so peers can
// reclaim un-fired wheel contents on the next tick instead of waiting
// out the assigned_until_utc lease. Only relinquishes timers whose
// claim is still held (assigned_until_utc in the future) to avoid
// stomping a peer that already reclaimed them after our lease expired.
var execBulkRelinquish = fmt.Sprintf(`UPDATE %s
SET
	assigned_worker = NULL
	, assigned_until_utc = NULL
	, attempt = GREATEST(attempt - 1, 0)
WHERE
	id = ANY($1)
	AND assigned_worker = $2
	AND assigned_until_utc IS NOT NULL
	AND assigned_until_utc > $3
	AND delivered_utc IS NULL
`, timerTableName)

func (m Manager) BulkRelinquish(ctx context.Context, workerIdentity string, asOf time.Time, ids []uuid.UUID) (err error) {
	if len(ids) == 0 {
		return
	}
	_, err = m.bulkRelinquish.ExecContext(ctx, ids, workerIdentity, asOf)
	return
}

var execDeleteTimerByID = fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, timerTableName)

func (m Manager) DeleteTimerByID(ctx context.Context, id uuid.UUID) (found bool, err error) {
	var res sql.Result
	res, err = m.deleteTimerByID.ExecContext(ctx, id)
	if err != nil {
		return
	}
	rows, _ := res.RowsAffected()
	found = rows > 0
	return
}

var execDeleteTimerByName = fmt.Sprintf(`DELETE FROM %s WHERE name = $1`, timerTableName)

func (m Manager) DeleteTimerByName(ctx context.Context, name string) (found bool, err error) {
	var res sql.Result
	res, err = m.deleteTimerByName.ExecContext(ctx, name)
	if err != nil {
		return
	}
	rows, _ := res.RowsAffected()
	found = rows > 0
	return
}

func (m Manager) DeleteTimers(ctx context.Context, after, before time.Time, matchLabels map[string]string) error {
	var args []any
	stanzas := []string{
		fmt.Sprintf("DELETE FROM %s WHERE 1=1", timerTableName),
	}
	if !after.IsZero() {
		args = append(args, after)
		stanzas = append(stanzas, fmt.Sprintf("AND due_utc > $%d", len(args)))
	}
	if !before.IsZero() {
		args = append(args, before)
		stanzas = append(stanzas, fmt.Sprintf("AND due_utc < $%d", len(args)))
	}
	for key, value := range matchLabels {
		args = append(args, key)
		args = append(args, value)
		stanzas = append(stanzas, fmt.Sprintf("AND labels->$%d = $%d", len(args)-1, len(args)))
	}

	statement := strings.Join(stanzas, "\n")
	_, err := m.Invoke(ctx).Exec(statement, args...)
	return err
}

const execWorkerSeen = `INSERT INTO workers (hostname, created_utc, last_seen_utc)
VALUES ($1, $2, $2) ON CONFLICT (hostname) DO UPDATE SET last_seen_utc = $2`

func (m Manager) WorkerSeen(ctx context.Context, workerHostname string, ts time.Time) (err error) {
	_, err = m.workerSeen.ExecContext(ctx, workerHostname, ts)
	return
}

const execDeleteWorker = `DELETE FROM workers WHERE hostname = $1`

// DeleteWorker removes a worker's row from the workers table. Called on
// graceful shutdown so the departing worker's shard band is reclaimed
// by peers on the next tick instead of waiting out the staleness window.
func (m Manager) DeleteWorker(ctx context.Context, hostname string) error {
	_, err := m.deleteWorker.ExecContext(ctx, hostname)
	return err
}

var queryGetWorkers = fmt.Sprintf(`SELECT %s FROM workers WHERE last_seen_utc > $1`, db.ColumnNamesCSV(workerColumns))

func (m Manager) GetWorkers(ctx context.Context, asOf time.Time) (output []Worker, err error) {
	var rows *sql.Rows
	rows, err = m.getWorkers.QueryContext(ctx, asOf)
	if err != nil {
		return
	}
	for rows.Next() {
		var t Worker
		if err = db.PopulateInOrder(&t, rows, workerColumns); err != nil {
			return
		}
		output = append(output, t)
	}
	return
}

//
// batch ops
//

const execBulkMarkDelivered = `UPDATE timers SET delivered_utc = $1 WHERE id = ANY($2)`

func (m Manager) BulkMarkDelivered(ctx context.Context, deliveredUTC time.Time, ids []uuid.UUID) (err error) {
	_, err = m.bulkMarkDelivered.ExecContext(ctx, deliveredUTC, ids)
	return
}

var queryGetPeakTimersDueCount = fmt.Sprintf(`SELECT COALESCE(MAX(cnt), 0) FROM (
	SELECT count(*) as cnt
	FROM %s
	WHERE due_utc >= $1 AND due_utc < $2
		AND delivered_utc IS NULL
		AND attempt < 5
	GROUP BY floor(extract(epoch from due_utc) / $3)
) sub`, timerTableName)

func (m Manager) GetPeakTimersDueCount(ctx context.Context, after, before time.Time, bucketSeconds float64) (count int64, err error) {
	err = m.getPeakTimersDueCount.QueryRowContext(ctx, after, before, bucketSeconds).Scan(&count)
	return
}

var queryGetOverdueTimerCount = fmt.Sprintf(`SELECT count(*)
FROM %s
WHERE due_utc < $1
	AND delivered_utc IS NULL
	AND attempt < 5
`, timerTableName)

func (m Manager) GetOverdueTimerCount(ctx context.Context, asOf time.Time) (count int64, err error) {
	err = m.getOverdueTimerCount.QueryRowContext(ctx, asOf).Scan(&count)
	return
}
