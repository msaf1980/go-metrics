package metrics

import (
	"sync"
)

// Rate hold an int64 value and timestamp (current and previous) and return diff and diff/s.
type Rate interface {
	Snapshot() Rate
	Clear() (float64, float64)
	Update(v float64, timestamp_ns int64)
	Values() (float64, float64)
}

// GetOrRegisterRate returns an existing Rate or constructs and registers a
// new StandardRate.
func GetOrRegisterRate(name string, r Registry) Rate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewRate).(Rate)
}

// GetOrRegisterRateT returns an existing Rate or constructs and registers a
// new StandardRate.
func GetOrRegisterRateT(name string, tagsMap map[string]string, r Registry) Rate {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewRate).(Rate)
}

// NewRate constructs a new StandardRate.
func NewRate() Rate {
	if UseNilMetrics {
		return NilRate{}
	}
	return &StandardRate{}
}

// NewRegisteredRate constructs and registers a new StandardRate.
func NewRegisteredRate(name string, r Registry) Rate {
	c := NewRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredRateT constructs and registers a new StandardRate.
func NewRegisteredRateT(name string, tagsMap map[string]string, r Registry) Rate {
	c := NewRate()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// RateSnapshot is a read-only copy of another Rate.
type RateSnapshot struct {
	value float64
	rate  float64
}

func newRateSnapshot(v, rate float64) Rate {
	return &RateSnapshot{value: v, rate: rate}
}

// Snapshot returns the snapshot.
func (g *RateSnapshot) Snapshot() Rate { return g }

// Clear panics.
func (RateSnapshot) Clear() (float64, float64) {
	panic("Clear called on a RateSnapshot")
}

// Update panics.
func (RateSnapshot) Update(float64, int64) {
	panic("Update called on a RateSnapshot")
}

// Value returns the value at the time the snapshot was taken.
func (g *RateSnapshot) Values() (float64, float64) { return g.value, g.rate }

// NilRate is a no-op Rate.
type NilRate struct{}

// Snapshot is a no-op.
func (NilRate) Snapshot() Rate { return NilRate{} }

// Clear is a no-op.
func (NilRate) Clear() (float64, float64) { return 0, 0 }

// Update is a no-op.
func (NilRate) Update(float64, int64) {}

// Value is a no-op.
func (NilRate) Values() (float64, float64) { return 0, 0 }

// StandardRate is the standard implementation of a Rate and uses the
// sync/atomic package to manage a single int64 value.
type StandardRate struct {
	prev    float64
	prevTs  int64
	value   float64
	valueTs int64
	mutex   sync.Mutex
}

// Snapshot returns a read-only copy of the Rate.
func (g *StandardRate) Snapshot() Rate {
	return newRateSnapshot(g.Values())
}

// Clear sets the DownCounter to zero.
func (g *StandardRate) Clear() (float64, float64) {
	var (
		v float64
		d float64
	)
	g.mutex.Lock()
	if g.prevTs > 0 {
		v = g.value - g.prev
		d = float64(g.valueTs - g.prevTs)
	}
	g.value = 0
	g.prev = 0
	g.valueTs = 0
	g.prevTs = 0
	g.mutex.Unlock()
	if v <= 0 {
		// broken values
		return 0, 0
	}
	return v, 1e9 * v / d
}

// Update updates the Rate's value.
func (g *StandardRate) Update(v float64, ts int64) {
	g.mutex.Lock()
	g.prev = g.value
	g.prevTs = g.valueTs
	g.value = v
	g.valueTs = ts
	g.mutex.Unlock()
}

// Value returns the Rate's current value.
func (g *StandardRate) Values() (float64, float64) {
	var (
		v float64
		d float64
	)
	g.mutex.Lock()
	if g.prevTs > 0 {
		v = g.value - g.prev
		d = float64(g.valueTs - g.prevTs)
	}
	g.mutex.Unlock()
	if d == 0 {
		// first or immediate try
		return 0, 0
	}
	return v, 1e9 * v / d
}
