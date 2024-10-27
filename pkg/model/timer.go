package model

import (
	"time"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/uuid"
)

var _ db.TableNameProvider = (*Timer)(nil)

var timerTypeMeta = db.TypeMetaFor(Timer{})
var timerTableName = db.TableName(Timer{})
var timerColumns = timerTypeMeta.Columns()

// Timer is a promise in the future to deliver an RPC
type Timer struct {
	ID       uuid.UUID         `db:"id,pk,auto"`
	Name     string            `db:"name"`
	Labels   map[string]string `db:"labels,json"`
	Priority uint32            `db:"priority"`

	CreatedUTC time.Time `db:"created_utc"`
	DueUTC     time.Time `db:"due_utc"`

	DueCounter     uint64  `db:"due_counter"`
	RetryCounter   uint32  `db:"retry_counter"`
	AttemptCounter uint32  `db:"attempt_counter"`
	Attempt        uint32  `db:"attempt"`
	AssignedWorker *string `db:"assigned_worker"`

	HookURL     string            `db:"hook_url"`
	HookMethod  string            `db:"hook_method"`
	HookHeaders map[string]string `db:"hook_headers,json"`
	HookBody    []byte            `db:"hook_body"`

	DeliveredUTC        *time.Time `db:"delivered_utc"`
	DeliveredStatusCode uint32     `db:"delivered_status_code"`
	DeliveredErr        string     `db:"delivered_err"`
}

func (t Timer) MatchLabels() map[string]string {
	output := make(map[string]string, len(t.Labels))
	for key, value := range t.Labels {
		output[key] = value
	}
	if t.AssignedWorker != nil {
		output["assigned"] = "true"
		output["assigned_worker"] = *t.AssignedWorker
	}
	if t.DeliveredUTC != nil && !t.DeliveredUTC.IsZero() {
		output["delivered"] = "true"
	}
	return output
}

// TableName returns the table name.
func (t Timer) TableName() string { return "timers" }
