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
	deleteTimerByID     *sql.Stmt
	deleteTimerByName   *sql.Stmt
	bulkMarkDelivered   *sql.Stmt
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
	if err := m.bulkMarkDelivered.Close(); err != nil {
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

var queryGetDueTimers = fmt.Sprintf(`UPDATE %[1]s
SET 
	assigned_worker = $1
	, attempt = attempt + 1
	, assigned_until_utc = $2::timestamp + interval '1 minute'
	, retry_utc = $2 + interval '5 minutes'
WHERE
	id in (
		SELECT 
			id
		FROM
			%[1]s
		WHERE
			due_utc < $2 
			AND (
				assigned_until_utc IS NULL OR (assigned_until_utc IS NOT NULL AND assigned_until_utc < $2) 
			)
			AND (
				retry_utc IS NULL OR (retry_utc IS NOT NULL AND retry_utc < $2) 
			)
			AND attempt < 5
			AND delivered_utc IS NULL
		ORDER BY 
			priority + (
				(
					MOD(shard, 3600) + cast(extract('minute', $2::timestamp) * extract('second', $2::timestamp) as BIGINT)
				)
				* 100
			)
		DESC
		LIMIT $3
		FOR UPDATE SKIP LOCKED
)
RETURNING %[2]s
`, timerTableName, db.ColumnNamesCSV(timerColumns))

//

func (m Manager) GetDueTimers(ctx context.Context, workerIdentity string, asOf time.Time, batchSize int) (output []Timer, err error) {
	var rows *sql.Rows
	rows, err = m.getDueTimers.QueryContext(ctx, workerIdentity, asOf, batchSize)
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

func (m Manager) CullTimers(ctx context.Context, cutoff time.Time) (err error) {
	_, err = m.cullTimers.ExecContext(ctx, cutoff)
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

//
// batch ops
//

const execBulkMarkDelivered = `UPDATE timers SET delivered_utc = $1 WHERE id = ANY($2)`

func (m Manager) BulkMarkDelivered(ctx context.Context, deliveredUTC time.Time, ids []uuid.UUID) (err error) {
	_, err = m.bulkMarkDelivered.ExecContext(ctx, deliveredUTC, ids)
	return
}
