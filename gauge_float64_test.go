package metrics

import "testing"

func BenchmarkGuageFloat64(b *testing.B) {
	g := NewFGauge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(float64(i))
	}
}

func BenchmarkGuageFloat64Parallel(b *testing.B) {
	g := NewFGauge()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Update(float64(1))
		}
	})
}

func TestFGauge(t *testing.T) {
	g := NewFGauge()
	g.Update(float64(47.0))
	if v := g.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestFGaugeSnapshot(t *testing.T) {
	g := NewFGauge()
	g.Update(float64(47.0))
	snapshot := g.Snapshot()
	g.Update(float64(0))
	if v := snapshot.Value(); float64(47.0) != v {
		t.Errorf("g.Value(): 47.0 != %v\n", v)
	}
}

func TestGetOrRegisterFGauge(t *testing.T) {
	r := NewRegistry()
	NewRegisteredFGauge("foo", r).Update(float64(47.0))
	// t.Logf("registry: %v", r)
	if g := GetOrRegisterFGauge("foo", r); float64(47.0) != g.Value() {
		t.Fatal(g)
	}
}

func TestFunctionalFGauge(t *testing.T) {
	var counter float64
	fg := NewFunctionalFGauge(func() float64 {
		counter++
		return counter
	})
	fg.Value()
	fg.Value()
	if counter != 2 {
		t.Error("counter != 2")
	}
}

func TestGetOrRegisterFunctionalFGauge(t *testing.T) {
	r := NewRegistry()
	NewRegisteredFunctionalFGauge("foo", r, func() float64 { return 47 })
	if g := GetOrRegisterFGauge("foo", r); g.Value() != 47 {
		t.Fatal(g)
	}
}
