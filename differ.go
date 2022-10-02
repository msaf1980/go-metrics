package metrics

import (
	"sync"
)

// GetOrRegisterDiffer returns an existing Differ or constructs and registers a
// new StandardDiffer.
func GetOrRegisterDiffer(name string, r Registry) Gauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewDiffer).(Gauge)
}

// GetOrRegisterDifferT returns an existing Differ or constructs and registers a
// new StandardDiffer.
func GetOrRegisterDifferT(name, tags string, r Registry) Gauge {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tags, NewDiffer).(Gauge)
}

// NewDiffer constructs a new StandardDiffer.
func NewDiffer() Gauge {
	if UseNilMetrics {
		return NilGauge{}
	}
	return &StandardDiffer{}
}

// NewRegisteredDiffer constructs and registers a new StandardDiffer.
func NewRegisteredDiffer(name string, r Registry) Gauge {
	c := NewDiffer()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredDifferT constructs and registers a new StandardDiffer.
func NewRegisteredDifferT(name, tags string, r Registry) Gauge {
	c := NewDiffer()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tags, c)
	return c
}

// StandardDiffer is the standard implementation of a Differ and uses the
// sync/atomic package to manage a single int64 value.
type StandardDiffer struct {
	prev  int64
	value int64
	mutex sync.Mutex
}

// Snapshot returns a read-only copy of the Differ.
func (g *StandardDiffer) Snapshot() Gauge {
	return GaugeSnapshot(g.Value())
}

// Update updates the Differ's value.
func (g *StandardDiffer) Update(v int64) {
	g.mutex.Lock()
	g.prev = g.value
	g.value = v
	g.mutex.Unlock()
}

// Value returns the Differ's current value.
func (g *StandardDiffer) Value() int64 {
	g.mutex.Lock()
	v := g.value - g.prev
	g.mutex.Unlock()
	return v
}
