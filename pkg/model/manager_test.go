package model

import (
	"context"
	"fmt"
	"sandman/pkg/utils"
	"slices"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
	"go.charczuk.com/sdk/uuid"
)

func Test_Manager_GetDueTimers_byDueUTC(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-00",
		DueUTC: now.Add(time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-01",
		DueUTC: now.Add(2 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-02",
		DueUTC: now.Add(2 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
		Attempt:    5,
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-03",
		DueUTC: now.Add(4 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now.Add(3*time.Hour), 10)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}

func Test_Manager_GetDueTimers_byAssignedUntilUTC(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:             "test-timer-00",
		DueUTC:           now,
		AssignedUntilUTC: utils.Ref(now.Add(time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:             "test-timer-01",
		DueUTC:           now,
		AssignedUntilUTC: utils.Ref(now.Add(2 * time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:             "test-timer-02",
		DueUTC:           now,
		AssignedUntilUTC: utils.Ref(now.Add(2 * time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
		Attempt:    5,
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:             "test-timer-03",
		DueUTC:           now,
		AssignedUntilUTC: utils.Ref(now.Add(4 * time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now.Add(3*time.Hour), 10)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}

func Test_Manager_GetDueTimers_byRetryUTC(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:     "test-timer-00",
		DueUTC:   now,
		RetryUTC: utils.Ref(now.Add(time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)
	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:     "test-timer-01",
		DueUTC:   now,
		RetryUTC: utils.Ref(now.Add(2 * time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:     "test-timer-02",
		DueUTC:   now,
		RetryUTC: utils.Ref(now.Add(4 * time.Hour)),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now.Add(3*time.Hour), 10)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}

func Test_Manager_GetDueTimers_ordersByShard(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	t00 := Timer{
		Name:   "test-timer-00",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t00)
	assert.ItsNil(t, err)

	t01 := Timer{
		Name:   "test-timer-01",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t01)
	assert.ItsNil(t, err)

	t02 := Timer{
		Name:   "test-timer-02",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_also_not_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t02)
	assert.ItsNil(t, err)

	t03 := Timer{
		Name:   "test-timer-03",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_aswell_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t03)
	assert.ItsNil(t, err)

	t04 := Timer{
		Name:   "test-timer-04",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_sortof_aswell_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t04)
	assert.ItsNil(t, err)

	asOf := now.Add(3 * time.Hour)
	pseudoPriorities := []pseudoPriority{
		computePseudoPriority(asOf, &t00),
		computePseudoPriority(asOf, &t01),
		computePseudoPriority(asOf, &t02),
		computePseudoPriority(asOf, &t03),
		computePseudoPriority(asOf, &t04),
	}

	slices.SortFunc(pseudoPriorities, func(i, j pseudoPriority) int {
		if i.Priority < j.Priority {
			return 1
		} else if i.Priority == j.Priority {
			return 0
		}
		return -1
	})

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", asOf, 3)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 3, len(timers))

	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t01.ID) })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t02.ID) })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t03.ID) })

	assert.ItsEqual(t, t01.ID, pseudoPriorities[0].ID)
	assert.ItsEqual(t, t02.ID, pseudoPriorities[2].ID)
	assert.ItsEqual(t, t03.ID, pseudoPriorities[1].ID)
}

type pseudoPriority struct {
	ID       uuid.UUID
	Priority uint32
}

func computePseudoPriority(asOf time.Time, t *Timer) (output pseudoPriority) {
	output.ID = t.ID
	output.Priority += t.Priority

	// the idea here is we split each hour into 60 minutes and 60 seconds
	// for 3600 total shards.
	// we then assign the shard to a "bucket" based on the current minute and second, and offset
	// the bucket by the shard assignment onto 3600.
	// we then multiply by 100 giving us 0 -> 36,0000 possible "boost" to
	// the priority based on the shuffle shard
	bucket := uint32(asOf.Minute()) * uint32(asOf.Second())
	shardWeight := ((t.Shard % 3600) + bucket) % 3600
	output.Priority += shardWeight * 100
	return
}

func Test_Manager_GetDueTimers_ordersByShard_boosted(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	t00 := Timer{
		Name:   "test-timer-00",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t00)
	assert.ItsNil(t, err)

	t01 := Timer{
		Name:   "test-timer-01",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t01)
	assert.ItsNil(t, err)

	t02 := Timer{
		Name:   "test-timer-02",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_also_not_bufoco")),
		Priority:   400000, // this means it's beyond the 360,000 maximum boost of shuffling
	}
	err = modelMgr.Invoke(ctx).Create(&t02)
	assert.ItsNil(t, err)

	t03 := Timer{
		Name:   "test-timer-03",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_aswell_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t03)
	assert.ItsNil(t, err)

	t04 := Timer{
		Name:   "test-timer-04",
		DueUTC: now,
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		Shard:      StableHash([]byte("uk_not_sortof_aswell_bufoco")),
		Priority:   10,
	}
	err = modelMgr.Invoke(ctx).Create(&t04)
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", now.Add(3*time.Hour), 3)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 3, len(timers))

	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t01.ID) })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t02.ID) })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t03.ID) })
}

func Test_Manager_BulkMarkDelivered(t *testing.T) {
	ctx := context.Background()
	tx, err := testutil.DefaultDB().BeginTx(ctx)
	assert.ItsNil(t, err)
	defer tx.Rollback()

	modelMgr := &Manager{
		BaseManager: dbutil.NewBaseManager(
			testutil.DefaultDB(),
			db.OptTx(tx),
		),
	}
	err = modelMgr.Initialize(ctx)
	assert.ItsNil(t, err)
	defer modelMgr.Close()

	now := time.Date(2024, 10, 19, 20, 19, 18, 17, time.UTC)

	timers := make([]Timer, 100)
	for x := 0; x < 100; x++ {
		timers[x] = Timer{
			Name:   fmt.Sprintf("test-timer-%d", x),
			DueUTC: now.Add(time.Hour),
			Labels: map[string]string{
				"service": "test-service",
				"env":     "prod",
				"region":  "us-east-1",
			},
			CreatedUTC: now,
		}
		err = modelMgr.Invoke(ctx).Create(&timers[x])
		assert.ItsNil(t, err)
	}

	var ids = []uuid.UUID{
		timers[0].ID,
		timers[5].ID,
		timers[10].ID,
		timers[15].ID,
		timers[20].ID,
		timers[25].ID,
		timers[30].ID,
		timers[35].ID,
		timers[40].ID,
		timers[45].ID,
	}

	err = modelMgr.BulkMarkDelivered(ctx, now.Add(time.Hour), ids)
	assert.ItsNil(t, err)

	var verifyTimers []Timer
	err = modelMgr.Invoke(ctx).All(&verifyTimers)
	assert.ItsNil(t, err)

	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[0].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[1].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[2].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[3].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[4].ID) && t.DeliveredUTC == nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[5].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[10].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[15].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[20].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[25].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[30].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[35].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[40].ID) && t.DeliveredUTC != nil })
	assert.ItsAny(t, verifyTimers, func(t Timer) bool { return t.ID.Equal(timers[45].ID) && t.DeliveredUTC != nil })
}
