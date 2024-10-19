package model

import (
	"time"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/uuid"
)

var _ db.TableNameProvider = (*Timer)(nil)

// Timer is a promise in the future to deliver an RPC
type Timer struct {
	ID              uuid.UUID         `yaml:"-" db:"id,pk,auto"`
	ShardID         uint32            `yaml:"-" db:"shard_id"`
	Name            string            `yaml:"name" db:"name"`
	Labels          map[string]string `yaml:"labels" db:"labels,json"`
	CreatedUTC      time.Time         `yaml:"-" db:"created_utc"`
	DueUTC          time.Time         `yaml:"due_utc" db:"due_utc"`
	Attempt         uint64            `yaml:"-" db:"attempt"`
	AssignableUTC   *time.Time        `yaml:"-" db:"assignable_utc"`
	AssignedWorker  *string           `yaml:"-" db:"assigned_worker"`
	RPCAddr         string            `yaml:"rpc_addr" db:"rpc_addr"`
	RPCAuthority    string            `yaml:"rpc_authority" db:"rpc_authority"`
	RPCMethod       string            `yaml:"rpc_method" db:"rpc_method"`
	RPCMeta         map[string]string `yaml:"rpc_meta" db:"rpc_meta,json"`
	RPCArgs         []byte            `yaml:"rpc_args" db:"rpc_args"`
	DeliveredUTC    *time.Time        `yaml:"-" db:"delivered_utc"`
	DeliveredStatus uint32            `yaml:"-" db:"delivered_status"`
	DeliveredErr    string            `yaml:"-" db:"delivered_err"`
}

// TableName returns the table name.
func (t Timer) TableName() string { return "timers" }
