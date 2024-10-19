package migration

import (
	"fmt"
	"strings"

	"go.charczuk.com/sdk/log"
)

// NewEvent returns a new event.
func NewEvent(result, body string, labels ...string) *Event {
	return &Event{
		Result: result,
		Body:   body,
		Labels: labels,
	}
}

// Event is a migration logger event.
type Event struct {
	Result string
	Body   string
	Labels []string
}

// WriteText writes the migration event as text.
func (e Event) String() string {
	wr := new(strings.Builder)
	if len(e.Result) > 0 {
		wr.WriteString("--")
		wr.WriteString(" ")
		wr.WriteString(e.Result)
	}
	if len(e.Labels) > 0 {
		wr.WriteString(" ")
		wr.WriteString(strings.Join(e.Labels, " > "))
	}
	if len(e.Body) > 0 {
		wr.WriteString(" -- ")
		wr.WriteString(e.Body)
	}
	return wr.String()
}

// NewStatsEvent returns a new stats event.
func NewStatsEvent(applied, skipped, failed, total int) *StatsEvent {
	return &StatsEvent{
		applied: applied,
		skipped: skipped,
		failed:  failed,
		total:   total,
	}
}

// StatsEvent is a migration logger event.
type StatsEvent struct {
	applied int
	skipped int
	failed  int
	total   int
}

// WriteText writes the event to a text writer.
func (se StatsEvent) String() string {
	return fmt.Sprintf("%d applied %d skipped %d failed %d total", se.applied, se.skipped, se.failed, se.total)
}

// Attrs implements log.AttrProvider.
func (se StatsEvent) Attrs() []log.Attr {
	return []log.Attr{
		log.Int("applied", se.applied),
		log.Int("skipped", se.skipped),
		log.Int("failed", se.failed),
		log.Int("total", se.total),
	}
}
