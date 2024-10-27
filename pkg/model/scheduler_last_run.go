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
	Namespace  string     `db:"namespace,pk"`
	Worker     *string    `db:"worker"`
	LastRunUTC *time.Time `db:"last_run_utc"`
}

func (s SchedulerLastRun) TableName() string { return "scheduler_last_run" }
