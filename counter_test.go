package metrics

import "testing"

func BenchmarkCounter(b *testing.B) {
	c := NewCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Add(1)
	}
}

func BenchmarkCounterParallel(b *testing.B) {
	c := NewCounter()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(1)
		}
	})
}

func TestCounterClear(t *testing.T) {
	c := NewCounter()
	c.Add(1)
	c.Clear()
	if count := c.Count(); count != 0 {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestCounterInc1(t *testing.T) {
	c := NewCounter()
	c.Add(1)
	if count := c.Count(); count != 1 {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestCounterInc2(t *testing.T) {
	c := NewCounter()
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

func TestCounterSnapshot(t *testing.T) {
	c := NewCounter()
	c.Add(1)
	snapshot := c.Snapshot()
	c.Add(1)
	if count := snapshot.Count(); count != 1 {
		t.Errorf("c.Count(): 1 != %v\n", count)
	}
}

func TestCounterZero(t *testing.T) {
	c := NewCounter()
	if count := c.Count(); count != 0 {
		t.Errorf("c.Count(): 0 != %v\n", count)
	}
}

func TestGetOrRegisterCounter(t *testing.T) {
	r := NewRegistry()
	NewRegisteredCounter("foo", r).Add(47)
	if c := GetOrRegisterCounter("foo", r); c.Count() != 47 {
		t.Fatal(c)
	}
}
