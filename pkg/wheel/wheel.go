// Package wheel implements an in-memory single-level timer wheel used as
// a worker-side dispatch buffer in front of the durable timers table.
//
// Each slot holds the timers whose due_utc falls in a one-second window;
// Advance returns the timers whose slot the cursor has reached or
// passed, in time order, so the caller can fire them. Storage is the
// source of truth — the wheel is purely a cache of claimed timers and
// nothing here is durable across restarts.
package wheel

import (
	"sort"
	"sync"
	"time"

	"sandman/pkg/model"
	"sandman/pkg/uuid"
)

// New returns a wheel with slotCount one-second slots whose cursor is
// aligned to anchor truncated to the second. slotCount must be > 0;
// callers should size it strictly greater than the prefetch window so
// inserts never collide with the cursor mid-lap.
func New(slotCount int, anchor time.Time) *Wheel {
	if slotCount <= 0 {
		panic("wheel: slotCount must be > 0")
	}
	return &Wheel{
		slots:    make([]slot, slotCount),
		cursor:   0,
		cursorAt: anchor.UTC().Truncate(time.Second),
		ids:      make(map[uuid.UUID]struct{}),
	}
}

type Wheel struct {
	mu       sync.Mutex
	slots    []slot
	cursor   int
	cursorAt time.Time
	// ids tracks every timer currently held in any slot so callers can
	// dedupe across overlapping prefetches without a slot-by-slot scan.
	ids map[uuid.UUID]struct{}
}

type slot struct {
	timers []*model.Timer
}

// Size returns the number of slots in the wheel.
func (w *Wheel) Size() int { return len(w.slots) }

// CursorAt returns the wall-clock second the wheel's cursor currently
// points at. Useful in tests and for instrumentation.
func (w *Wheel) CursorAt() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.cursorAt
}

// Len returns how many timers the wheel is currently holding across all
// slots. Used by the worker to bound prefetch fills and surface as a
// metric.
func (w *Wheel) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.ids)
}

// Insert places t in the slot for its DueUTC. Timers already past due
// land in the cursor's slot so they fire on the next Advance. Returns
// false if t is already held by the wheel (e.g. a re-claim during an
// overlapping prefetch); the caller can drop the duplicate without
// touching the DB.
//
// If t.DueUTC is further in the future than the wheel can represent
// (>= slotCount seconds past the cursor) Insert returns false and the
// caller should leave the timer for a later prefetch — the lease will
// still expire safely if it is never picked back up.
func (w *Wheel) Insert(t *model.Timer) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.ids[t.ID]; ok {
		return false
	}
	offset := max(int(t.DueUTC.UTC().Truncate(time.Second).Sub(w.cursorAt)/time.Second), 0)
	if offset >= len(w.slots) {
		return false
	}
	idx := (w.cursor + offset) % len(w.slots)
	w.slots[idx].timers = append(w.slots[idx].timers, t)
	w.ids[t.ID] = struct{}{}
	return true
}

// Advance moves the cursor forward to now and returns every timer in
// any slot whose wall-clock second has been reached (inclusive of the
// slot at now). Timers are returned in due_utc order so the caller can
// dispatch newest-overdue last; callers that care about priority should
// re-sort the returned slice.
//
// Advance is safe to call more frequently than once per second; sub-
// second calls are idempotent because slots are emptied as they fire.
func (w *Wheel) Advance(now time.Time) []*model.Timer {
	w.mu.Lock()
	defer w.mu.Unlock()
	now = now.UTC().Truncate(time.Second)
	var fired []*model.Timer
	// Drain every slot whose wall-clock time is at or before now,
	// catching up if the dispatcher fell behind. Bounded by slotCount
	// so a long pause can't loop forever — anything older than that
	// has already been overwritten by a later prefetch's wrap-around.
	for steps := 0; steps < len(w.slots) && !w.cursorAt.After(now); steps++ {
		s := &w.slots[w.cursor]
		if len(s.timers) > 0 {
			for _, t := range s.timers {
				delete(w.ids, t.ID)
			}
			fired = append(fired, s.timers...)
			s.timers = nil
		}
		w.cursor = (w.cursor + 1) % len(w.slots)
		w.cursorAt = w.cursorAt.Add(time.Second)
	}
	if len(fired) > 1 {
		sort.SliceStable(fired, func(i, j int) bool {
			return fired[i].DueUTC.Before(fired[j].DueUTC)
		})
	}
	return fired
}

// DrainAll empties the wheel and returns everything it was holding,
// regardless of due_utc. Used during graceful shutdown so the worker
// can relinquish its claim on un-fired timers in one DB call.
func (w *Wheel) DrainAll() []*model.Timer {
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []*model.Timer
	for i := range w.slots {
		if len(w.slots[i].timers) == 0 {
			continue
		}
		out = append(out, w.slots[i].timers...)
		w.slots[i].timers = nil
	}
	w.ids = make(map[uuid.UUID]struct{})
	return out
}
