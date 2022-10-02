package metrics

import (
	"sync/atomic"
)

// Healthchecks hold an error value describing an arbitrary up/down status.
type Healthcheck interface {
	Check()
	IsUp() bool
	Healthy()
	Unhealthy()
}

// NewHealthcheck constructs a new Healthcheck which will use the given
// function to update its status.
func NewHealthcheck(f func(bool) bool) Healthcheck {
	if UseNilMetrics {
		return NilHealthcheck{}
	}
	return &StandardHealthcheck{f: f}
}

// NilHealthcheck is a no-op.
type NilHealthcheck struct{}

// Check is a no-op.
func (NilHealthcheck) Check() {}

// IsError is a no-op.
func (NilHealthcheck) IsUp() bool { return true }

// Healthy is a no-op.
func (NilHealthcheck) Healthy() {}

// Unhealthy is a no-op.
func (NilHealthcheck) Unhealthy() {}

// StandardHealthcheck is the standard implementation of a Healthcheck and
// stores the status and a function to call to update the status.
type StandardHealthcheck struct {
	up int32
	f  func(bool) bool
}

// Check runs the healthcheck function to update the healthcheck's status.
func (h *StandardHealthcheck) Check() {
	if up := h.f(h.IsUp()); up {
		atomic.CompareAndSwapInt32(&h.up, 0, 1)
	} else {
		atomic.CompareAndSwapInt32(&h.up, 1, 0)
	}
}

// IsUp returns the healthcheck's status
func (h *StandardHealthcheck) IsUp() bool {
	return atomic.LoadInt32(&h.up) > 0
}

// Healthy marks the healthcheck as healthy.
func (h *StandardHealthcheck) Healthy() {
	atomic.StoreInt32(&h.up, 1)
}

// Unhealthy marks the healthcheck as unhealthy.
func (h *StandardHealthcheck) Unhealthy() {
	atomic.StoreInt32(&h.up, 0)
}
