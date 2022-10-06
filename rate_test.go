package metrics

import (
	"fmt"
	"testing"
)

func BenchmarkRate(b *testing.B) {
	g := NewRate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.UpdateTs(int64(i), int64(i))
	}
}

func TestRate(t *testing.T) {
	g := NewRate()

	g.UpdateTs(1, 1e9)
	v, rate := g.Values()
	if v != 1 {
		t.Errorf("g.Values() first value: 0 != %v\n", v)
	}
	if rate != 0 {
		t.Errorf("g.Values() first value: 0 != %v\n", rate)
	}

	g.UpdateTs(7, 3e9)
	v, rate = g.Values()
	if v != 7 {
		t.Errorf("g.Values() alue: 7 != %v\n", v)
	}
	if rate != 3 {
		t.Errorf("g.Values() value: 3 != %v\n", rate)
	}

	v, rate = g.Clear()
	if v != 7 {
		t.Errorf("g.Clear()  value: 7 != %v\n", v)
	}
	if rate != 3 {
		t.Errorf("g.Clear() first value: 3 != %v\n", rate)
	}

	v, rate = g.Values()
	if v != 0 {
		t.Errorf("g.Values() clear value: 0 != %v\n", v)
	}
	if rate != 0 {
		t.Errorf("g.Values() clear value: 0 != %v\n", rate)
	}
}

func TestRateSnapshot(t *testing.T) {
	g := NewRate()
	g.UpdateTs(1, 1e9)
	g.UpdateTs(7, 3e9)

	snapshot := g.Snapshot()
	g.UpdateTs(8, 4e9)

	v, rate := snapshot.Values()
	if v != 7 {
		t.Errorf("g.Values() value: 6 != %v\n", v)
	}
	if rate != 3 {
		t.Errorf("g.Values() rate: 3 != %v\n", rate)
	}
}

func TestGetOrRegisterRate(t *testing.T) {
	r := NewRegistry()
	NewRegisteredRate("foo", r).UpdateTs(1, 1e9)
	GetOrRegisterRate("foo", r).UpdateTs(7, 3e9)
	v, rate := GetOrRegisterRate("foo", r).Values()
	if v != 7 {
		t.Errorf("g.Values() value: 7 != %v\n", v)
	}
	if rate != 3 {
		t.Errorf("g.Values() rate: 3 != %v\n", rate)
	}
}

func ExampleGetOrRegisterRate() {
	m := "server.memory_used"
	r := NewRegistry()
	g := GetOrRegisterRate(m, r)
	g.UpdateTs(1, 1e9)
	g.UpdateTs(7, 3e9)
	fmt.Println(g.Values()) // Output: 7 3
}
