package scheduler

import (
	"context"
	"expvar"
	"time"

	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

func New(identity string, mgr *model.Manager) *Scheduler {
	return &Scheduler{
		mgr: mgr,
	}
}

type Scheduler struct {
	identity              string
	mgr                   *model.Manager
	updateTickInterval    time.Duration
	heartbeatTickInterval time.Duration

	schedulerGeneration        expvar.Int
	schedulerTicks             expvar.Int
	schedulerTickErrors        expvar.Int
	schedulerElapsedLastMillis expvar.Int

	generation uint64
	isLeader   bool
}

type SchedulerVars struct {
	SchedulerGeneration        *expvar.Int
	SchedulerTicks             *expvar.Int
	SchedulerTickErrors        *expvar.Int
	SchedulerElapsedLastMillis *expvar.Int
}

func (sv SchedulerVars) Publish() {
	expvar.Publish("scheduler_generation", sv.SchedulerGeneration)
	expvar.Publish("scheduler_elapsed_last_millis", sv.SchedulerElapsedLastMillis)
	expvar.Publish("scheduler_ticks", sv.SchedulerTicks)
	expvar.Publish("scheduler_tick_errors", sv.SchedulerTickErrors)
}

func (s *Scheduler) Vars() SchedulerVars {
	return SchedulerVars{
		SchedulerGeneration:        &s.schedulerGeneration,
		SchedulerTicks:             &s.schedulerTicks,
		SchedulerTickErrors:        &s.schedulerTickErrors,
		SchedulerElapsedLastMillis: &s.schedulerElapsedLastMillis,
	}
}

const defaultUpdateTickInterval = time.Minute

func (s *Scheduler) updateTickIntervalOrDefault() time.Duration {
	if s.updateTickInterval > 0 {
		return s.updateTickInterval
	}
	return defaultUpdateTickInterval
}

const defaultHeartbeatTickInterval = time.Minute

func (s *Scheduler) heartbeatTickIntervalOrDefault() time.Duration {
	if s.heartbeatTickInterval > 0 {
		return s.heartbeatTickInterval
	}
	return defaultHeartbeatTickInterval
}

func (s *Scheduler) Run(ctx context.Context) error {
	var err error
	s.generation, s.isLeader, err = s.mgr.AttemptLeaderElection(ctx, s.identity, s.generation)
	if err != nil {
		return err
	}

	if !s.isLeader {
		log.GetLogger(ctx).Info("scheduler entering follower mode, polling for new leader")
		if err = s.awaitNextElection(ctx); err != nil {
			return err
		}
	}

	lastRun, err := s.mgr.GetLastRun(ctx)
	if err != nil {
		return err
	}
	if lastRun.IsZero() {
		now := time.Now().UTC().Add(time.Minute)
		delta := time.Duration(now.Nanosecond()) + (time.Duration(now.Second()) * time.Second)
		log.GetLogger(ctx).Info("scheduler sleeping", log.Duration("for", delta))
		s.sleepFor(ctx, time.Duration(delta))
	} else if delta := time.Now().UTC().Sub(lastRun.UTC()); delta < time.Minute {
		log.GetLogger(ctx).Info("scheduler sleeping", log.Duration("for", delta))
		s.sleepFor(ctx, delta)
	}

	// instantaneously kick off a pass
	deadlineCtx, deadlineCancel := context.WithTimeout(ctx, s.updateTickIntervalOrDefault())
	go func() {
		defer deadlineCancel()
		s.processTick(deadlineCtx)
	}()

	// start the tick loop to iterate
	// and process timers
	updateTick := time.NewTicker(s.updateTickIntervalOrDefault())
	defer updateTick.Stop()
	heartbeatTick := time.NewTicker(s.heartbeatTickIntervalOrDefault())
	defer updateTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-updateTick.C:
			deadlineCtx, deadlineCancel = context.WithTimeout(ctx, s.heartbeatTickIntervalOrDefault())
			go func() {
				defer deadlineCancel()
				s.processTick(deadlineCtx)
			}()
		case <-heartbeatTick.C:
			if err = s.mgr.Heartbeat(ctx); err != nil {
				return err
			}
		}
	}
}

func (s *Scheduler) awaitNextElection(ctx context.Context) error {
	ticker := time.NewTicker(s.heartbeatTickIntervalOrDefault())
	defer ticker.Stop()
	var err error
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.generation, s.isLeader, err = s.mgr.AttemptLeaderElection(ctx, s.identity, s.generation+1)
			if err != nil {
				return err
			}
			if s.isLeader {
				return nil
			}
		}
	}
}

func (w *Scheduler) sleepFor(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return
	case <-t.C:
		return
	}
}

func (w *Scheduler) processTick(ctx context.Context) {
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		log.GetLogger(ctx).Info("scheduler; updated timers", log.Duration("elapsed", elapsed))
		w.schedulerElapsedLastMillis.Set(int64(elapsed / time.Millisecond))
	}()
	w.schedulerTicks.Add(1)
	if err := w.mgr.UpdateTimers(ctx, time.Now().UTC()); err != nil {
		w.schedulerTickErrors.Add(1)
		log.GetLogger(ctx).Error("scheduler; failed to update timers", log.Any("err", err))
	}
}
