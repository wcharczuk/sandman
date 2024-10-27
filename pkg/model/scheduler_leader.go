package model

import (
	"time"

	"go.charczuk.com/sdk/db"
)

var _ db.TableNameProvider = (*Timer)(nil)

var (
	schedulerLeaderTypeMeta  = db.TypeMetaFor(SchedulerLeader{})
	schedulerLeaderTableName = db.TableName(SchedulerLeader{})
	schedulerLeaderColumns   = schedulerLeaderTypeMeta.Columns()
)

type SchedulerLeader struct {
	Generation  uint64     `db:"generation"`
	Leader      *string    `db:"leader"`
	LastSeenUTC *time.Time `db:"last_seen_utc"`
}

func (s SchedulerLeader) TableName() string { return "scheduler_leader" }
