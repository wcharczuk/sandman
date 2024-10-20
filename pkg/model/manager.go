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
	markDelivered       *sql.Stmt
	markAttempted       *sql.Stmt
	deleteTimerByID     *sql.Stmt
	deleteTimerByName   *sql.Stmt
}

func (m *Manager) Initialize(ctx context.Context) (err error) {
	m.getDueTimers, err = m.Invoke(ctx).Prepare(queryGetDueTimers)
	if err != nil {
		return
	}
	m.getTimerByName, err = m.Invoke(ctx).Prepare(queryGetTimerByName)
	if err != nil {
		return
	}
	m.getTimersDueBetween, err = m.Invoke(ctx).Prepare(queryGetTimersDueBetween)
	if err != nil {
		return
	}
	m.cullTimers, err = m.Invoke(ctx).Prepare(execCullTimers)
	if err != nil {
		return
	}
	m.markDelivered, err = m.Invoke(ctx).Prepare(execMarkDelivered)
	if err != nil {
		return
	}
	m.markAttempted, err = m.Invoke(ctx).Prepare(execMarkAttempted)
	if err != nil {
		return
	}
	m.deleteTimerByID, err = m.Invoke(ctx).Prepare(execDeleteTimerByID)
	if err != nil {
		return
	}
	m.deleteTimerByName, err = m.Invoke(ctx).Prepare(execDeleteTimerByName)
	if err != nil {
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

var queryGetDueTimers = fmt.Sprintf(`UPDATE
	%s
SET 
	assigned_worker = $1
	, attempt = attempt + 1
	, assignable_utc = $2 + interval '1 minute'
WHERE
	due_utc < $2
	AND attempt < 5
	AND (assignable_utc IS NULL OR assignable_utc < $2)
	AND delivered_utc IS NULL
RETURNING %s
`, timerTableName, db.ColumnNamesCSV(timerColumns))

func (m Manager) GetDueTimers(ctx context.Context, workerIdentity string, asOf time.Time) (output []Timer, err error) {
	var rows *sql.Rows
	rows, err = m.getDueTimers.QueryContext(ctx, workerIdentity, asOf)
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
	delivered_utc IS NOT NULL
	OR attempt >= 5
`, timerTableName)

func (m Manager) CullTimers(ctx context.Context) (err error) {
	_, err = m.cullTimers.ExecContext(ctx)
	return nil
}

var execMarkDelivered = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_utc = current_timestamp
	, assignable_utc = NULL
WHERE 
	id = $1
`, timerTableName)

func (m Manager) MarkDelivered(ctx context.Context, id uuid.UUID) (err error) {
	_, err = m.markDelivered.ExecContext(ctx, id)
	return
}

var execMarkAttempted = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_status = $2
	, delivered_err = $3
	, assignable_utc = NULL
WHERE 
	id = $1
`, timerTableName)

func (m Manager) MarkAttempted(ctx context.Context, id uuid.UUID, deliveredStatus uint32, deliveredErr string) (err error) {
	_, err = m.markAttempted.ExecContext(ctx, id, deliveredStatus, deliveredErr)
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
