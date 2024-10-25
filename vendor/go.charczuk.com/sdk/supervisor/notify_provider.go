package supervisor

import "github.com/rjeczalik/notify"

// FileEventProvider handles hooking up file event notifications.
type FileEventProvider interface {
	Notify(string, chan notify.EventInfo) error
}

// NotifyProvider is the concrete notify provider.
type NotifyProvider struct{}

// Notify calls the underlying notification implemention.
func (np NotifyProvider) Notify(watchedPath string, fsevents chan notify.EventInfo) error {
	return notify.Watch(watchedPath, fsevents, notify.All)
}

// MockNotifyProvider is a mocked notify provider.
type MockNotifyProvider struct {
	watchedPaths []string
	events       chan notify.EventInfo
}

// Notify calls the underlying notification implemention.
func (mnp *MockNotifyProvider) Notify(watchedPath string, fsevents chan notify.EventInfo) error {
	mnp.watchedPaths = append(mnp.watchedPaths, watchedPath)
	mnp.events = fsevents
	return nil
}

// Signal sends an event to the channel.
func (mnp *MockNotifyProvider) Signal(event notify.EventInfo) {
	mnp.events <- event
}
