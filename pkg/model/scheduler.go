package model

import (
	"time"

	"go.charczuk.com/sdk/db"
)

var _ db.TableNameProvider = (*Timer)(nil)

var schedulerTypeMeta = db.TypeMetaFor(Scheduler{})
var schedulerTableName = db.TableName(Scheduler{})
var schedulerColumns = schedulerTypeMeta.Columns()

type Scheduler struct {
	Name    string    `db:"name,pk"`
	LastRun time.Time `db:"last_run"`
}

func (s Scheduler) TableName() string { return "scheduler" }
