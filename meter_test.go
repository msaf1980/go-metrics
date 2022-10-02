package metrics

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/msaf1980/go-metrics/test"
)

func BenchmarkMeter(b *testing.B) {
	m := NewMeter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Mark(1)
	}
}

func BenchmarkMeterParallel(b *testing.B) {
	m := NewMeter()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Mark(1)
		}
	})
}

// exercise race detector
func TestMeterConcurrency(t *testing.T) {
	rand.Seed(time.Now().Unix())
	ma := meterArbiter{
		ticker: time.NewTicker(time.Millisecond),
		meters: make(map[*StandardMeter]struct{}),
	}
	m := newStandardMeter()
	ma.meters[m] = struct{}{}
	go ma.tick()
	wg := &sync.WaitGroup{}
	reps := 100
	for i := 0; i < reps; i++ {
		wg.Add(1)
		go func(m Meter, wg *sync.WaitGroup) {
			m.Mark(1)
			wg.Done()
		}(m, wg)
		wg.Add(1)
		go func(m Meter, wg *sync.WaitGroup) {
			m.Stop()
			wg.Done()
		}(m, wg)
	}
	wg.Wait()
}

func TestGetOrRegisterMeter(t *testing.T) {
	r := NewRegistry()
	NewRegisteredMeter("foo", r).Mark(47)
	if m := GetOrRegisterMeter("foo", r); m.Count() != 47 {
		t.Fatal(m)
	}
}

func TestMeterDecay(t *testing.T) {
	ma := meterArbiter{
		ticker: time.NewTicker(time.Millisecond),
		meters: make(map[*StandardMeter]struct{}),
	}
	m := newStandardMeter()
	ma.meters[m] = struct{}{}
	go ma.tick()
	m.Mark(1)
	rateMean := m.RateMean()
	time.Sleep(100 * time.Millisecond)
	if m.RateMean() >= rateMean {
		t.Error("m.RateMean() didn't decrease")
	}
}

func TestMeterNonzero(t *testing.T) {
	m := NewMeter()
	m.Mark(3)
	if count := m.Count(); count != 3 {
		t.Errorf("m.Count(): 3 != %v\n", count)
	}
}

func TestMeterStop(t *testing.T) {
	l := len(arbiter.meters)
	m := NewMeter()
	if len(arbiter.meters) != l+1 {
		t.Errorf("arbiter.meters: %d != %d\n", l+1, len(arbiter.meters))
	}
	m.Stop()
	if len(arbiter.meters) != l {
		t.Errorf("arbiter.meters: %d != %d\n", l, len(arbiter.meters))
	}
}

func TestMeterSnapshot(t *testing.T) {
	m := NewMeter()
	m.Mark(1)
	if snapshot := m.Snapshot(); m.RateMean() != snapshot.RateMean() || m.Count() != snapshot.Count() {
		t.Fatal(snapshot)
	}
}

func TestMeterZero(t *testing.T) {
	m := NewMeter()
	if count := m.Count(); count != 0 {
		t.Errorf("m.Count(): 0 != %v\n", count)
	}
}

func TestMeter(t *testing.T) {
	rand.Seed(time.Now().Unix())
	ma := meterArbiter{
		ticker: time.NewTicker(time.Millisecond),
		meters: make(map[*StandardMeter]struct{}),
	}
	m := newStandardMeter()
	ma.meters[m] = struct{}{}
	go ma.tick()

	time.Sleep(10 * time.Millisecond)

	m.Mark(47)

	time.Sleep(10 * time.Millisecond)

	if want, v := int64(47), m.Count(); v != want {
		t.Errorf("metric.Count() = %d, want %d", v, want)
	}
	if want, v := 469.01, m.RateMean(); !test.FloatEqDev(v, want, 10) {
		t.Errorf("metric.RateMean() = %f, want %f", v, want)
	}
	if want, v := 9.4, m.Rate1(); !test.FloatEq(v, want) {
		t.Errorf("metric.Rate1() = %f, want %f", v, want)
	}
	if want, v := 9.4, m.Rate5(); !test.FloatEq(v, want) {
		t.Errorf("metric.Rate5() = %f, want %f", v, want)
	}
	if want, v := 9.4, m.Rate15(); !test.FloatEq(v, want) {
		t.Errorf("metric.Rate15() = %f, want %f", v, want)
	}
}
