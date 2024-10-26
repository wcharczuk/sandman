package model

import (
	"context"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
)

func Test_Manager_GetDueTimers(t *testing.T) {
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
		DueCounter: 0,
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
		DueCounter: 0,
	})
	assert.ItsNil(t, err)

	err = modelMgr.Invoke(ctx).Create(&Timer{
		Name:   "test-timer-02",
		DueUTC: now.Add(4 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		DueCounter: 1,
	})
	assert.ItsNil(t, err)

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker")
	assert.ItsNil(t, err)
	assert.ItsEqual(t, 2, len(timers))
}

func Test_Manager_UpdateTimers(t *testing.T) {
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

	t0 := Timer{
		Name:   "test-timer-00",
		DueUTC: now.Add(time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC:   now,
		DueCounter:   0,
		RetryCounter: 2,
		Attempt:      1,
	}
	err = modelMgr.Invoke(ctx).Create(&t0)
	assert.ItsNil(t, err)

	t1 := Timer{
		Name:   "test-timer-01",
		DueUTC: now.Add(2 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC:     now,
		DueCounter:     0,
		AttemptCounter: 3,
		Attempt:        1,
	}
	err = modelMgr.Invoke(ctx).Create(&t1)
	assert.ItsNil(t, err)

	t2 := Timer{
		Name:   "test-timer-02",
		DueUTC: now.Add(4 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		DueCounter: 4,
		Attempt:    0,
	}
	err = modelMgr.Invoke(ctx).Create(&t2)
	assert.ItsNil(t, err)

	err = modelMgr.UpdateTimers(ctx, time.Now().UTC(), 1)
	assert.ItsNil(t, err)

	var timers []Timer
	err = modelMgr.Invoke(ctx).All(&timers)
	assert.ItsNil(t, err)

	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t0.ID) && t.RetryCounter == 1 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t1.ID) && t.AttemptCounter == 2 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t2.ID) && t.DueCounter == 3 })
}

func Test_Manager_UpdateTimers_multiMinuteUpdate(t *testing.T) {
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

	t0 := Timer{
		Name:   "test-timer-00",
		DueUTC: now.Add(time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC:   now,
		DueCounter:   0,
		RetryCounter: 2,
		Attempt:      1,
	}
	err = modelMgr.Invoke(ctx).Create(&t0)
	assert.ItsNil(t, err)

	t1 := Timer{
		Name:   "test-timer-01",
		DueUTC: now.Add(2 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC:     now,
		DueCounter:     0,
		AttemptCounter: 3,
		Attempt:        1,
	}
	err = modelMgr.Invoke(ctx).Create(&t1)
	assert.ItsNil(t, err)

	t2 := Timer{
		Name:   "test-timer-02",
		DueUTC: now.Add(4 * time.Hour),
		Labels: map[string]string{
			"service": "test-service",
			"env":     "prod",
			"region":  "us-east-1",
		},
		CreatedUTC: now,
		DueCounter: 4,
		Attempt:    0,
	}
	err = modelMgr.Invoke(ctx).Create(&t2)
	assert.ItsNil(t, err)

	err = modelMgr.UpdateTimers(ctx, time.Now().UTC(), 5)
	assert.ItsNil(t, err)

	var timers []Timer
	err = modelMgr.Invoke(ctx).All(&timers)
	assert.ItsNil(t, err)

	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t0.ID) && t.RetryCounter == 0 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t1.ID) && t.AttemptCounter == 0 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t2.ID) && t.DueCounter == 0 })
}
