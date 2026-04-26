package model

import (
	"time"

	"sandman/pkg/db"
)

var (
	_               db.TableNameProvider = (*Worker)(nil)
	workerTypeMeta                       = db.TypeMetaFor(Worker{})
	workerTableName                      = db.TableName(Worker{})
	workerColumns                        = workerTypeMeta.Columns()
)

type Worker struct {
	Hostname    string    `db:"hostname,pk"`
	CreatedUTC  time.Time `db:"created_utc"`
	LastSeenUTC time.Time `db:"last_seen_utc"`
}

// TableName returns the table name.
func (w Worker) TableName() string { return "workers" }
