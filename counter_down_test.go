package metrics

import "testing"

func BenchmarkDownCounter(b *testing.B) {
	c := NewDownCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(1)
	}
}

func BenchmarkDownCounterParallel(b *testing.B) {
	c := NewDownCounter()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
}

func TestDownCounterClear(t *testing.T) {
	c := NewDownCounter()
	c.Add(1)
	c.Clear()
	if count := c.Count(); count != 0 {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestDownCounterDec1(t *testing.T) {
	c := NewDownCounter()
	c.Sub(1)
	if count := c.Count(); count != -1 {
		t.Errorf("c.Count(): -1 != %v\n", count)
	}
}

func TestDownCounterDec2(t *testing.T) {
	c := NewDownCounter()
	c.Sub(2)
	if count := c.Count(); count != -2 {
		t.Errorf("c.Count(): -2 != %v\n", count)
	}
}

func TestDownCounterInc1(t *testing.T) {
	c := NewDownCounter()
	c.Add(1)
	if count := c.Count(); count != 1 {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestDownCounterInc2(t *testing.T) {
	c := NewDownCounter()
	c.Add(2)
	if count := c.Count(); count != 2 {
		t.Errorf("c.Count(): 2 != %v\n", count)
	}
	n := c.Clear()
	if n != 2 {
		t.Errorf("c.Clear(): 2 != %v\n", n)
	}
	if count := c.Count(); count != 0 {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestDownCounterSnapshot(t *testing.T) {
	c := NewDownCounter()
	c.Add(1)
	snapshot := c.Snapshot()
	c.Add(1)
	if count := snapshot.Count(); count != 1 {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestDownCounterZero(t *testing.T) {
	c := NewDownCounter()
	if count := c.Count(); count != 0 {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestGetOrRegisterDownCounter(t *testing.T) {
	r := NewRegistry()
	NewRegisteredDownCounter("foo", r).Add(47)
	if c := GetOrRegisterDownCounter("foo", r); c.Count() != 47 {
		t.Fatal(c)
	}
}
