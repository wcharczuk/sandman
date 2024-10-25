package scheduler

import (
	"context"
	"expvar"
	"time"

	"go.charczuk.com/sdk/log"

	"sandman/pkg/model"
)

func New(mgr *model.Manager) *Scheduler {
	return &Scheduler{
		mgr: mgr,
	}
}

type Scheduler struct {
	mgr          *model.Manager
	tickInterval time.Duration

	schedulerTicks             expvar.Int
	schedulerTickErrors        expvar.Int
	schedulerElapsedLastMillis expvar.Int
}

type SchedulerVars struct {
	SchedulerTicks             *expvar.Int
	SchedulerTickErrors        *expvar.Int
	SchedulerElapsedLastMillis *expvar.Int
}

func (sv SchedulerVars) Publish() {
	expvar.Publish("scheduler_elapsed_last_millis", sv.SchedulerElapsedLastMillis)
	expvar.Publish("scheduler_ticks", sv.SchedulerTicks)
	expvar.Publish("scheduler_tick_errors", sv.SchedulerTickErrors)
}

func (s *Scheduler) Vars() SchedulerVars {
	return SchedulerVars{
		SchedulerTicks:             &s.schedulerTicks,
		SchedulerTickErrors:        &s.schedulerTickErrors,
		SchedulerElapsedLastMillis: &s.schedulerElapsedLastMillis,
	}
}

const defaultTickInterval = time.Minute

func (s *Scheduler) tickIntervalOrDefault() time.Duration {
	if s.tickInterval > 0 {
		return s.tickInterval
	}
	return defaultTickInterval
}

func (s *Scheduler) Run(ctx context.Context) error {
	tick := time.NewTicker(s.tickIntervalOrDefault())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			deadlineCtx, deadlineCancel := context.WithTimeout(ctx, s.tickIntervalOrDefault())
			go func() {
				defer deadlineCancel()
				s.processTick(deadlineCtx)
			}()
		}
	}
}

func (w *Scheduler) processTick(ctx context.Context) {
	started := time.Now()
	defer func() {
		elapsed := time.Since(started)
		log.GetLogger(ctx).Error("scheduler; updated timers", log.Duration("elapsed", elapsed))
		w.schedulerElapsedLastMillis.Set(int64(elapsed / time.Millisecond))
	}()
	w.schedulerTicks.Add(1)
	if err := w.mgr.UpdateTimers(ctx, time.Now().UTC()); err != nil {
		w.schedulerTickErrors.Add(1)
		log.GetLogger(ctx).Error("scheduler; failed to update timers", log.Any("err", err))
	}
}