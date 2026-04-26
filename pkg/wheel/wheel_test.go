package wheel

import (
	"testing"
	"time"

	"sandman/pkg/uuid"

	"sandman/pkg/model"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	out, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return out
}

func newTimer(due time.Time) *model.Timer {
	return &model.Timer{ID: uuid.V4(), DueUTC: due}
}

func TestInsertAndAdvance_FiresInSlot(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(64, now)

	a := newTimer(now.Add(2 * time.Second))
	b := newTimer(now.Add(5 * time.Second))
	if !w.Insert(a) || !w.Insert(b) {
		t.Fatal("expected both inserts to succeed")
	}
	if got := w.Len(); got != 2 {
		t.Fatalf("Len: got %d want 2", got)
	}

	if fired := w.Advance(now.Add(1 * time.Second)); len(fired) != 0 {
		t.Fatalf("advance to t+1: got %d fired, want 0", len(fired))
	}
	fired := w.Advance(now.Add(2 * time.Second))
	if len(fired) != 1 || fired[0].ID != a.ID {
		t.Fatalf("advance to t+2: want a, got %+v", fired)
	}
	fired = w.Advance(now.Add(5 * time.Second))
	if len(fired) != 1 || fired[0].ID != b.ID {
		t.Fatalf("advance to t+5: want b, got %+v", fired)
	}
	if got := w.Len(); got != 0 {
		t.Fatalf("Len after drain: got %d want 0", got)
	}
}

func TestAdvance_DrainsCatchUpInOrder(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(64, now)

	timers := []*model.Timer{
		newTimer(now.Add(3 * time.Second)),
		newTimer(now.Add(1 * time.Second)),
		newTimer(now.Add(5 * time.Second)),
		newTimer(now.Add(2 * time.Second)),
	}
	for _, tm := range timers {
		if !w.Insert(tm) {
			t.Fatalf("insert failed for %v", tm.DueUTC)
		}
	}

	fired := w.Advance(now.Add(10 * time.Second))
	if len(fired) != 4 {
		t.Fatalf("got %d fired, want 4", len(fired))
	}
	for i := 1; i < len(fired); i++ {
		if fired[i-1].DueUTC.After(fired[i].DueUTC) {
			t.Fatalf("not sorted by due_utc: %v then %v", fired[i-1].DueUTC, fired[i].DueUTC)
		}
	}
}

func TestInsert_OverdueLandsInCursorSlot(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(64, now)

	overdue := newTimer(now.Add(-30 * time.Second))
	if !w.Insert(overdue) {
		t.Fatal("expected overdue insert to succeed")
	}

	fired := w.Advance(now)
	if len(fired) != 1 || fired[0].ID != overdue.ID {
		t.Fatalf("expected overdue fired immediately, got %+v", fired)
	}
}

func TestInsert_BeyondHorizonRejected(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(8, now)

	tooFar := newTimer(now.Add(30 * time.Second))
	if w.Insert(tooFar) {
		t.Fatal("expected insert beyond horizon to be rejected")
	}
	if got := w.Len(); got != 0 {
		t.Fatalf("Len: got %d want 0", got)
	}
}

func TestInsert_DuplicateIDRejected(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(64, now)

	tm := newTimer(now.Add(2 * time.Second))
	if !w.Insert(tm) {
		t.Fatal("first insert failed")
	}
	if w.Insert(tm) {
		t.Fatal("duplicate insert should have returned false")
	}
	if got := w.Len(); got != 1 {
		t.Fatalf("Len: got %d want 1", got)
	}
}

func TestAdvance_SubSecondDoesNotFireFutureTimer(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(64, now)

	tm := newTimer(now.Add(2 * time.Second))
	w.Insert(tm)

	// Sub-second wall-clock advances may drain the empty slot at the
	// current second, but must never fire a timer that lives further
	// in the future.
	if fired := w.Advance(now.Add(500 * time.Millisecond)); len(fired) != 0 {
		t.Fatalf("sub-second advance fired future timer: got %d, want 0", len(fired))
	}
	if fired := w.Advance(now.Add(1500 * time.Millisecond)); len(fired) != 0 {
		t.Fatalf("advance to t+1.5s fired t+2 timer: got %d, want 0", len(fired))
	}
	fired := w.Advance(now.Add(2 * time.Second))
	if len(fired) != 1 || fired[0].ID != tm.ID {
		t.Fatalf("advance to t+2 should fire the timer, got %+v", fired)
	}
}

func TestAdvance_LapWrapReusesSlots(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(4, now)

	// Fire one in slot 1, advance past the lap, then insert another
	// timer destined for what is now slot 1 again.
	a := newTimer(now.Add(1 * time.Second))
	w.Insert(a)
	if got := len(w.Advance(now.Add(2 * time.Second))); got != 1 {
		t.Fatalf("first lap: got %d fired, want 1", got)
	}

	// Cursor is now at t+2, slotCount=4 → can insert up to t+5.
	b := newTimer(now.Add(5 * time.Second))
	if !w.Insert(b) {
		t.Fatal("expected re-use insert to succeed")
	}
	fired := w.Advance(now.Add(5 * time.Second))
	if len(fired) != 1 || fired[0].ID != b.ID {
		t.Fatalf("expected b to fire after wrap, got %+v", fired)
	}
}

func TestDrainAll_ReturnsEverythingAndClears(t *testing.T) {
	now := mustTime(t, "2026-04-25T12:00:00Z")
	w := New(16, now)

	a := newTimer(now.Add(2 * time.Second))
	b := newTimer(now.Add(8 * time.Second))
	w.Insert(a)
	w.Insert(b)

	out := w.DrainAll()
	if len(out) != 2 {
		t.Fatalf("drain: got %d, want 2", len(out))
	}
	if w.Len() != 0 {
		t.Fatalf("Len after drain: got %d, want 0", w.Len())
	}
	// After drain, the IDs should be insertable again.
	if !w.Insert(a) {
		t.Fatal("expected re-insert after drain to succeed")
	}
}
