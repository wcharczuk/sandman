package model

import (
	"time"

	"go.charczuk.com/sdk/db"
)

var _ db.TableNameProvider = (*Timer)(nil)

var schedulerTypeMeta = db.TypeMetaFor(Scheduler{})
var schedulerTableName = db.TableName(Scheduler{})
var schedulerColumns = schedulerTypeMeta.Columns()

// Scheduler holds singleton information about the schedulers
// that is currently actively pushing timers forward.
type Scheduler struct {
	Name    string    `db:"name,pk"`
	LastRun time.Time `db:"last_run"`

	// Leader election stuff
	LastSeenUTC time.Time `db:"last_seen_utc"`
	Generation  uint64    `db:"generation"`
	Leader      string    `db:"leader"`
}

func (s Scheduler) TableName() string { return "scheduler" }
