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
	getLastRun          *sql.Stmt
	updateLastRun       *sql.Stmt
	updateTimers        *sql.Stmt
	cullTimers          *sql.Stmt
	markDelivered       *sql.Stmt
	markAttempted       *sql.Stmt
	deleteTimerByID     *sql.Stmt
	deleteTimerByName   *sql.Stmt
}

func (m *Manager) Initialize(ctx context.Context) (err error) {
	m.getDueTimers, err = m.Invoke(ctx).Prepare(queryGetDueTimers)
	if err != nil {
		err = fmt.Errorf("queryGetDueTimers: %w", err)
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
	m.getLastRun, err = m.Invoke(ctx).Prepare(queryGetLastRun)
	if err != nil {
		err = fmt.Errorf("getLastRun: %w", err)
		return
	}
	m.updateLastRun, err = m.Invoke(ctx).Prepare(execUpdateLastRun)
	if err != nil {
		err = fmt.Errorf("updateLastRun: %w", err)
		return
	}
	m.updateTimers, err = m.Invoke(ctx).Prepare(execUpdateTimers)
	if err != nil {
		err = fmt.Errorf("updateTimers: %w", err)
		return
	}
	m.cullTimers, err = m.Invoke(ctx).Prepare(execCullTimers)
	if err != nil {
		err = fmt.Errorf("cullTimers: %w", err)
		return
	}
	m.markDelivered, err = m.Invoke(ctx).Prepare(execMarkDelivered)
	if err != nil {
		err = fmt.Errorf("markDelivered: %w", err)
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
	return
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

// queryGetDueTimers is the query to poll for "due" timers
//
// when it marks timers for attempts it also advances the assignable time a minute into the future
var queryGetDueTimers = fmt.Sprintf(`UPDATE
	%s
SET 
	assigned_worker = $1
	, attempt = attempt + 1
	, retry_counter = 5
WHERE
	due_counter = 0
	AND attempt_counter = 0
	AND retry_counter = 0
	AND attempt < 5
	AND delivered_utc IS NULL
RETURNING %s
`, timerTableName, db.ColumnNamesCSV(timerColumns))

func (m Manager) GetDueTimers(ctx context.Context, workerIdentity string) (output []Timer, err error) {
	var rows *sql.Rows
	rows, err = m.getDueTimers.QueryContext(ctx, workerIdentity)
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

var queryGetLastRun = fmt.Sprintf(`SELECT last_run FROM %s WHERE name = $1`, schedulerTableName)

func (m Manager) GetLastRun(ctx context.Context) (lastRun time.Time, err error) {
	res, err := m.getLastRun.QueryContext(ctx, "default")
	if err != nil {
		err = fmt.Errorf("getLastRun: %w", err)
		return
	}
	if !res.Next() {
		return
	}
	err = res.Scan(&lastRun)
	return
}

var execUpdateLastRun = fmt.Sprintf(`INSERT INTO %s (name, last_run) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET last_run = $2`, schedulerTableName)

func (m Manager) UpdateLastRun(ctx context.Context, asOf time.Time) (err error) {
	_, err = m.updateLastRun.ExecContext(ctx, "default", asOf)
	if err != nil {
		err = fmt.Errorf("updateLastRun: %w", err)
	}
	return
}

var execUpdateTimers = fmt.Sprintf(`UPDATE %s
SET
	due_counter = case when due_counter > 0 then due_counter - 1 else 0 end
	, attempt_counter = case when attempt_counter > 0 then attempt_counter - 1 else 0 end
	, retry_counter = case when retry_counter > 0 then retry_counter - 1 else 0 end
WHERE 
	delivered_utc IS NOT NULL
	OR attempt >= 5
`, timerTableName)

func (m Manager) UpdateTimers(ctx context.Context, asOf time.Time) (err error) {
	_, err = m.updateTimers.ExecContext(ctx)
	if err == nil {
		err = m.UpdateLastRun(ctx, asOf)
	}
	return
}

var execCullTimers = fmt.Sprintf(`DELETE FROM %s 
WHERE 
	delivered_utc IS NOT NULL
	OR attempt >= 5
`, timerTableName)

func (m Manager) CullTimers(ctx context.Context) (err error) {
	_, err = m.cullTimers.ExecContext(ctx)
	return
}

var execMarkDelivered = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_utc = current_timestamp
WHERE 
	id = $1
`, timerTableName)

func (m Manager) MarkDelivered(ctx context.Context, id uuid.UUID) (err error) {
	_, err = m.markDelivered.ExecContext(ctx, id)
	return
}

var execMarkAttempted = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_status_code = $2
	, delivered_err = $3
	, retry_counter = 5
	, attempt_counter = 0
WHERE 
	id = $1
	AND attempt < 5
`, timerTableName)

func (m Manager) MarkAttempted(ctx context.Context, id uuid.UUID, deliveredStatus uint32, deliveredErr error) (err error) {
	var deliveredErrString string
	if deliveredErr != nil {
		deliveredErrString = deliveredErr.Error()
	}
	_, err = m.markAttempted.ExecContext(ctx, id, deliveredStatus, deliveredErrString)
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
