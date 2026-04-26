package async

import "sync/atomic"

// LatchThin is an implementation of a subset of the latch api
// that does not support notifying on channels.
//
// As a result, it's much easier to embed and use as a zero value.
type LatchThin struct {
	state int32
}

// Reset resets the latch.
func (lt *LatchThin) Reset() {
	atomic.StoreInt32(&lt.state, LatchStopped)
}

// CanStart returns if the latch can start.
func (lt LatchThin) CanStart() bool {
	return atomic.LoadInt32(&lt.state) == LatchStopped
}

// CanStop returns if the latch can stop.
func (lt LatchThin) CanStop() bool {
	return atomic.LoadInt32(&lt.state) == LatchStarted
}

// IsStarting returns if the latch state is LatchStarting
func (lt LatchThin) IsStarting() bool {
	return atomic.LoadInt32(&lt.state) == LatchStarting
}

// IsStarted returns if the latch state is LatchStarted.
func (lt LatchThin) IsStarted() bool {
	return atomic.LoadInt32(&lt.state) == LatchStarted
}

// IsStopping returns if the latch state is LatchStopping.
func (lt LatchThin) IsStopping() bool {
	return atomic.LoadInt32(&lt.state) == LatchStopping
}

// IsStopped returns if the latch state is LatchStopped.
func (lt LatchThin) IsStopped() bool {
	return atomic.LoadInt32(&lt.state) == LatchStopped
}

// Starting signals the latch is starting.
func (lt *LatchThin) Starting() bool {
	return atomic.CompareAndSwapInt32(&lt.state, LatchStopped, LatchStarting)
}

// Started signals that the latch is started and has entered the `IsStarted` state.
func (lt *LatchThin) Started() bool {
	return atomic.CompareAndSwapInt32(&lt.state, LatchStarting, LatchStarted)
}

// Stopping signals the latch to stop.
func (lt *LatchThin) Stopping() bool {
	return atomic.CompareAndSwapInt32(&lt.state, LatchStarted, LatchStopping)
}

// Stopped signals the latch has stopped.
func (lt *LatchThin) Stopped() bool {
	return atomic.CompareAndSwapInt32(&lt.state, LatchStopping, LatchStopped)
}
