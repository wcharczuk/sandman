package model

import (
	"context"
	"fmt"
	"sandman/pkg/utils"
	"testing"
	"time"

	"go.charczuk.com/sdk/assert"
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/dbutil"
	"go.charczuk.com/sdk/testutil"
	"go.charczuk.com/sdk/uuid"
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

	timers, err := modelMgr.GetDueTimers(ctx, "test-worker", 10)
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

	err = modelMgr.UpdateTimers(ctx, "test-worker", time.Now().UTC(), 1)
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

	err = modelMgr.UpdateTimers(ctx, "test-worker", time.Now().UTC(), 5)
	assert.ItsNil(t, err)

	var timers []Timer
	err = modelMgr.Invoke(ctx).All(&timers)
	assert.ItsNil(t, err)

	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t0.ID) && t.RetryCounter == 0 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t1.ID) && t.AttemptCounter == 0 })
	assert.ItsAny(t, timers, func(t Timer) bool { return t.ID.Equal(t2.ID) && t.DueCounter == 0 })
}

func Test_Manager_SchedulerLeaderElection_initial(t *testing.T) {
	/*
		Assert that the leader election works on a "fresh" state where
		we don't have a leader heartbeat time, and the generation is 0
	*/

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

	createTestNamespace(t, tx, NamespaceTesting)

	leaderTimeout := 30 * time.Second
	now := time.Date(2024, 10, 27, 19, 18, 17, 16, time.UTC)

	generation, isLeader, err := modelMgr.SchedulerLeaderElection(ctx, NamespaceTesting, "test-worker", 0, now, leaderTimeout)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, true, isLeader)
	assert.ItsEqual(t, 1, generation)
}

func Test_Manager_SchedulerLeaderElection_initial_existingLeader(t *testing.T) {
	/*
		Assert that the leader election works on a "valid" state where
		we do have a leader heartbeat time (which is not expired), and the generation is 5
	*/

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

	createTestNamespace(t, tx, NamespaceTesting)

	leaderTimeout := 30 * time.Second
	now := time.Date(2024, 10, 27, 19, 18, 17, 16, time.UTC)
	lastSeen := now.Add(-10 * time.Second)

	_ = modelMgr.Invoke(ctx).Upsert(&SchedulerLeader{
		Namespace:   NamespaceTesting,
		Generation:  5,
		Leader:      utils.Ref("not-test-worker"),
		LastSeenUTC: utils.Ref(lastSeen),
	})

	generation, isLeader, err := modelMgr.SchedulerLeaderElection(ctx, NamespaceTesting, "test-worker", 0, now, leaderTimeout)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, false, isLeader)
	assert.ItsEqual(t, 5, generation)
}

func Test_Manager_SchedulerLeaderElection_initial_existingLeader_timedout(t *testing.T) {
	/*
		Assert that the leader election works on a "timedout" state where
		we do have a leader heartbeat time (which is expired), and the generation is 5
	*/

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

	createTestNamespace(t, tx, NamespaceTesting)

	leaderTimeout := 30 * time.Second
	now := time.Date(2024, 10, 27, 19, 18, 17, 0, time.UTC)
	lastSeen := now.Add(-(leaderTimeout * 2))

	_ = modelMgr.Invoke(ctx).Upsert(&SchedulerLeader{
		Namespace:   NamespaceTesting,
		Generation:  5,
		Leader:      utils.Ref("not-test-worker"),
		LastSeenUTC: utils.Ref(lastSeen),
	})

	generation, isLeader, err := modelMgr.SchedulerLeaderElection(ctx, NamespaceTesting, "test-worker", 5, now, 30*time.Second)
	assert.ItsNil(t, err)
	assert.ItsEqual(t, true, isLeader)
	assert.ItsEqual(t, 6, generation)
}

func Test_Manager_BulkUpdateTimerSuccesses(t *testing.T) {
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
			DueCounter: 0,
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

	err = modelMgr.BulkUpdateTimerSuccesses(ctx, now.Add(time.Hour), ids)
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
