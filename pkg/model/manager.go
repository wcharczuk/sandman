package model

import (
	"context"
	"database/sql"
	"fmt"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/uuid"
)

type Manager struct {
	dbutil.BaseManager

	getDueTimers   *sql.Stmt
	getTimerByName *sql.Stmt
	cullTimers     *sql.Stmt
	markDelivered  *sql.Stmt
	markAttempted  *sql.Stmt
}

func (m Manager) Initialize(ctx context.Context) (err error) {
	m.getDueTimers, err = m.Invoke(ctx).Prepare(queryGetDueTimers)
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
	return
}

func (m Manager) CreateTimer(ctx context.Context, t *Timer, opts ...db.InvocationOption) (err error) {
	err = m.Invoke(ctx, opts...).Create(t)
	return
}

var timerTypeMeta = db.TypeMetaFor(Timer{})
var timerTableName = db.TableName(Timer{})
var timerColumns = timerTypeMeta.Columns()

var queryGetDueTimers = fmt.Sprintf(`SELECT
	%s
FROM
	%s AS t
WHERE
	t.shard_id = $1
	AND t.due_utc < current_timestamp
	AND t.attempt < 5
	AND (t.assignable_utc IS NULL OR t.assignable_utc < current_timestamp)
	AND t.delivered_utc IS NULL
FOR UPDATE SET 
	t.assigned = true
	, t.assigned_worker = $2
	, t.attempt = t.attempt + 1
	, t.assignable_utc = current_timestamp + interval '1 minute'
`, db.ColumnNamesFromAliasCSV(timerColumns, "t"), timerTableName)

func (m Manager) GetDueTimers(ctx context.Context, shardID uint32, workerIdentity string, opts ...db.InvocationOption) (output []Timer, err error) {
	var rows *sql.Rows
	rows, err = m.getDueTimers.QueryContext(ctx, shardID, workerIdentity)
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
	shard_id = $1 
	AND (
		delivered = true
		OR attempt >= 5
	)
`, timerTableName)

func (m Manager) CullTimers(ctx context.Context, shardID uint32) (err error) {
	_, err = m.cullTimers.ExecContext(ctx, shardID)
	return nil
}

var execMarkDelivered = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_utc = current_timestamp
	, assignable_utc = NULL
WHERE 
	shard_id = $1
	AND id = $1
`, timerTableName)

func (m Manager) MarkDelivered(ctx context.Context, shardID uint32, id uuid.UUID) (err error) {
	_, err = m.markDelivered.ExecContext(ctx, shardID, id)
	return
}

var execMarkAttempted = fmt.Sprintf(`UPDATE %s 
SET 
	delivered_status = $3
	, delivered_err = $4
	, assignable_utc = NULL
WHERE 
	shard_id = $1
	AND id = $1
`, timerTableName)

func (m Manager) MarkAttempted(ctx context.Context, shardID uint32, id uuid.UUID, deliveredStatus uint32, deliveredErr string) (err error) {
	_, err = m.markAttempted.ExecContext(ctx, shardID, id, deliveredStatus, deliveredErr)
	return
}
