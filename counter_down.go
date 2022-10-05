package metrics

import "sync/atomic"

// DownCounters hold an int64 value that can be incremented/decremented
type DownCounter interface {
	Clear()
	Count() int64
	Inc(int64)
	Dec(int64)
	Snapshot() DownCounter
}

// GetOrRegisterDownCounter returns an existing DownCounter or constructs and registers
// a new StandardDownCounter.
func GetOrRegisterDownCounter(name string, r Registry) DownCounter {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, NewDownCounter).(DownCounter)
}

// GetOrRegisterDownCounterT returns an existing DownCounter or constructs and registers
// a new StandardDownCounter.
func GetOrRegisterDownCounterT(name string, tagsMap map[string]string, r Registry) DownCounter {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, NewDownCounter).(DownCounter)
}

// NewDownCounter constructs a new StandardDownCounter.
func NewDownCounter() DownCounter {
	if UseNilMetrics {
		return NilDownCounter{}
	}
	return &StandardDownCounter{0}
}

// NewRegisteredDownCounter constructs and registers a new StandardDownCounter.
func NewRegisteredDownCounter(name string, r Registry) DownCounter {
	c := NewDownCounter()
	if nil == r {
		r = DefaultRegistry
	}
	r.Register(name, c)
	return c
}

// NewRegisteredDownCounterT constructs and registers a new StandardDownCounter.
func NewRegisteredDownCounterT(name string, tagsMap map[string]string, r Registry) DownCounter {
	c := NewDownCounter()
	if nil == r {
		r = DefaultRegistry
	}
	r.RegisterT(name, tagsMap, c)
	return c
}

// DownCounterSnapshot is a read-only copy of another DownCounter.
type DownCounterSnapshot uint64

// Clear panics.
func (DownCounterSnapshot) Clear() {
	panic("Clear called on a DownCounterSnapshot")
}

// Count returns the count at the time the snapshot was taken.
func (c DownCounterSnapshot) Count() int64 { return int64(c) }

// Inc panics.
func (DownCounterSnapshot) Inc(int64) {
	panic("Inc called on a DownCounterSnapshot")
}

// Inc panics.
func (DownCounterSnapshot) Dec(int64) {
	panic("Inc called on a DownCounterSnapshot")
}

// Snapshot returns the snapshot.
func (c DownCounterSnapshot) Snapshot() DownCounter { return c }

// NilDownCounter is a no-op DownCounter.
type NilDownCounter struct{}

// Clear is a no-op.
func (NilDownCounter) Clear() {}

// Count is a no-op.
func (NilDownCounter) Count() int64 { return 0 }

// Inc is a no-op.
func (NilDownCounter) Inc(i int64) {}

func (NilDownCounter) Dec(i int64) {}

// Snapshot is a no-op.
func (NilDownCounter) Snapshot() DownCounter { return NilDownCounter{} }

// StandardDownCounter is the standard implementation of a DownCounter and uses the
// sync/atomic package to manage a single uint64 value.
type StandardDownCounter struct {
	count int64
}

// Clear sets the DownCounter to zero.
func (c *StandardDownCounter) Clear() {
	atomic.StoreInt64(&c.count, 0)
}

// Count returns the current count.
func (c *StandardDownCounter) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

// Inc increments the DownCounter by the given amount.
func (c *StandardDownCounter) Inc(i int64) {
	atomic.AddInt64(&c.count, i)
}

// Dec decrements the DownCounter by the given amount.
func (c *StandardDownCounter) Dec(i int64) {
	atomic.AddInt64(&c.count, -i)
}

// Snapshot returns a read-only copy of the DownCounter.
func (c *StandardDownCounter) Snapshot() DownCounter {
	return DownCounterSnapshot(c.Count())
}
