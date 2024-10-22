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
	ID               uuid.UUID         `db:"id,pk,auto"`
	Name             string            `db:"name"`
	Labels           map[string]string `db:"labels,json"`
	CreatedUTC       time.Time         `db:"created_utc"`
	DueUTC           time.Time         `db:"due_utc"`
	Attempt          uint64            `db:"attempt"`
	AssignableUTC    *time.Time        `db:"assignable_utc"`
	AssignedWorker   *string           `db:"assigned_worker"`
	RPCAddr          string            `db:"rpc_addr"`
	RPCAuthority     string            `db:"rpc_authority"`
	RPCMethod        string            `db:"rpc_method"`
	RPCMeta          map[string]string `db:"rpc_meta,json"`
	RPCArgsTypeURL   string            `db:"rpc_args_type_url"`
	RPCArgsData      []byte            `db:"rpc_args_data"`
	RPCReturnTypeURL string            `db:"rpc_return_type_url"`
	DeliveredUTC     *time.Time        `db:"delivered_utc"`
	DeliveredStatus  uint32            `db:"delivered_status"`
	DeliveredErr     string            `db:"delivered_err"`
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
