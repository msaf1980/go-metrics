package metrics

import (
	"fmt"
	"testing"
)

func BenchmarkFRate(b *testing.B) {
	g := NewFRate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.UpdateTs(float64(i), int64(i))
	}
}

func TestFRate(t *testing.T) {
	g := NewFRate()

	g.UpdateTs(1, 1e9)
	v, FRate := g.Values()
	if v != 1 {
		t.Errorf("g.Values() first value: 1 != %v\n", v)
	}
	if FRate != 0 {
		t.Errorf("g.Values() first value: 0 != %v\n", FRate)
	}

	g.UpdateTs(7, 3e9)
	v, FRate = g.Values()
	if v != 7 {
		t.Errorf("g.Values() alue: 7 != %v\n", v)
	}
	if FRate != 3 {
		t.Errorf("g.Values() value: 3 != %v\n", FRate)
	}

	v, FRate = g.Clear()
	if v != 7 {
		t.Errorf("g.Clear()  value: 7 != %v\n", v)
	}
	if FRate != 3 {
		t.Errorf("g.Clear() first value: 3 != %v\n", FRate)
	}

	v, FRate = g.Values()
	if v != 0 {
		t.Errorf("g.Values() clear value: 0 != %v\n", v)
	}
	if FRate != 0 {
		t.Errorf("g.Values() clear value: 0 != %v\n", FRate)
	}
}

func TestFRateSnapshot(t *testing.T) {
	g := NewFRate()
	g.UpdateTs(1, 1e9)
	g.UpdateTs(7, 3e9)

	snapshot := g.Snapshot()
	g.UpdateTs(8, 4e9)

	v, FRate := snapshot.Values()
	if v != 7 {
		t.Errorf("g.Values() value: 7 != %v\n", v)
	}
	if FRate != 3 {
		t.Errorf("g.Values() FRate: 3 != %v\n", FRate)
	}
}

func TestGetOrRegisterFRate(t *testing.T) {
	r := NewRegistry()
	NewRegisteredFRate("foo", r).UpdateTs(1.2, 1e9)
	GetOrRegisterFRate("foo", r).UpdateTs(7.2, 3e9)
	v, rate := GetOrRegisterFRate("foo", r).Values()
	if v != 7.2 {
		t.Errorf("g.Values() value: 7.2 != %v\n", v)
	}
	if rate != 3 {
		t.Errorf("g.Values() FRate: 3 != %v\n", rate)
	}
}

func ExampleGetOrRegisterFRate() {
	m := "server.memory_used"
	r := NewRegistry()
	g := GetOrRegisterFRate(m, r)
	g.UpdateTs(1.1, 1e9)
	g.UpdateTs(7.1, 3e9)
	fmt.Println(g.Values()) // Output: 7.1 3
}
