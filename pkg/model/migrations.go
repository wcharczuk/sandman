package model

import (
	"go.charczuk.com/sdk/db/dbgen"
	"go.charczuk.com/sdk/db/migration"
)

func Migrations(opts ...migration.SuiteOption) *migration.Suite {
	return migration.New(
		append(opts,
			migration.OptGroups(
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						Timer{},
						dbgen.UniqueKey(Timer{}, "name"),
					),
				),
				migration.NewGroupWithAction(
					dbgen.TableFrom(
						Worker{},
						dbgen.Index(Worker{}, "last_seen_utc"),
					),
				),
				// The partial index is keyed (shard, due_utc) rather than
				// (due_utc) so writes spread across shard-prefix ranges
				// from the first insert instead of piling onto the tail
				// of a monotonic timestamp. The SPLIT EVENLY + SCATTER
				// makes that distribution effective on day one; otherwise
				// the allocator has to wait for load-based splits, during
				// which every insert hammers one leaseholder.
				//
				// SkipTransaction because mikoshi DDL is async: CREATE
				// INDEX returns after the descriptor is written but the
				// index isn't PUBLIC until the backfill job finishes.
				// A follow-up SPLIT in the same transaction hangs
				// waiting on a view of the index the txn will never see.
				migration.NewGroupWithStep(
					migration.IndexNotExists("timers", "ix_timers_shard_due_utc_pending"),
					migration.Statements(
						`CREATE INDEX ix_timers_shard_due_utc_pending ON timers (shard, due_utc) WHERE delivered_utc IS NULL AND attempt < 5`,
						`ALTER INDEX timers@ix_timers_shard_due_utc_pending SPLIT EVENLY FROM (0) TO (4294967296) INTO 16`,
						`ALTER INDEX timers@ix_timers_shard_due_utc_pending SCATTER`,
					),
					migration.OptGroupSkipTransaction(),
				),
				migration.NewGroupWithStep(
					migration.IndexExists("timers", "ix_timers_due_utc_pending"),
					migration.Statements(
						`DROP INDEX timers@ix_timers_due_utc_pending`,
					),
				),
				migration.NewGroupWithStep(
					migration.IndexExists("timers", "ix_timers_due_utc_attempt_delivered_utc_assigned_until_utc_retry_utc"),
					migration.Statements(
						`DROP INDEX timers@ix_timers_due_utc_attempt_delivered_utc_assigned_until_utc_retry_utc`,
					),
				),
			),
		)...,
	)
}
