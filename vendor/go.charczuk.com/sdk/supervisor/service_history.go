package supervisor

import "time"

// Exit is a time and a error associated with a service exiting.
type Exit struct {
	Timestamp time.Time
	Error     error
}

// ServiceHistory is the relevant bits of a service's history.
type ServiceHistory struct {
	StartedAt time.Time
	Exits     []Exit
}

// RecentFailures returns just the last N exits that have an error.
func (sh ServiceHistory) RecentFailures() (failures []Exit) {
	for x := len(sh.Exits) - 1; x >= 0; x-- {
		if sh.Exits[x].Error == nil {
			return
		}
		failures = append(failures, sh.Exits[x])
	}
	return
}
