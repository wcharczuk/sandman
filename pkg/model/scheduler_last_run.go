package model

import (
	"time"

	"go.charczuk.com/sdk/db"
)

var _ db.TableNameProvider = (*Timer)(nil)

var (
	schedulerLastRunTypeMeta  = db.TypeMetaFor(SchedulerLastRun{})
	schedulerLastRunTableName = db.TableName(SchedulerLastRun{})
	schedulerLastRunColumns   = schedulerLastRunTypeMeta.Columns()
)

// SchedulerLastRun holds singleton information about the schedulers
// that is currently actively pushing timers forward.
type SchedulerLastRun struct {
	LastRun time.Time `db:"last_run"`
}

func (s SchedulerLastRun) TableName() string { return "scheduler_last_run" }
