package metrics

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func BenchmarkUGauge(b *testing.B) {
	g := NewUGauge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(uint64(i))
	}
}

// exercise race detector
func TestUGaugeConcurrency(t *testing.T) {
	rand.Seed(time.Now().Unix())
	g := NewUGauge()
	wg := &sync.WaitGroup{}
	reps := 100
	for i := 0; i < reps; i++ {
		wg.Add(1)
		go func(g UGauge, wg *sync.WaitGroup) {
			g.Update(uint64(rand.Int63()))
			wg.Done()
		}(g, wg)
	}
	wg.Wait()
}

func TestUGauge(t *testing.T) {
	g := NewUGauge()
	g.Update(uint64(47))
	if v := g.Value(); v != 47 {
		t.Errorf("g.Value(): 47 != %v\n", v)
	}
}

func TestUGaugeSnapshot(t *testing.T) {
	g := NewUGauge()
	g.Update(uint64(47))
	snapshot := g.Snapshot()
	g.Update(uint64(0))
	if v := snapshot.Value(); v != 47 {
		t.Errorf("g.Value(): 47 != %v\n", v)
	}
}

func TestGetOrRegisterUGauge(t *testing.T) {
	r := NewRegistry()
	NewRegisteredUGauge("foo", r).Update(47)
	if g := GetOrRegisterUGauge("foo", r); g.Value() != 47 {
		t.Fatal(g)
	}
}

func TestFunctionalUGauge(t *testing.T) {
	var counter uint64
	fg := NewFunctionalUGauge(func() uint64 {
		counter++
		return counter
	})
	fg.Value()
	fg.Value()
	if counter != 2 {
		t.Error("counter != 2")
	}
}

func TestGetOrRegisterFunctionalUGauge(t *testing.T) {
	r := NewRegistry()
	NewRegisteredFunctionalUGauge("foo", r, func() uint64 { return 47 })
	if g := GetOrRegisterUGauge("foo", r); g.Value() != 47 {
		t.Fatal(g)
	}
}

func ExampleGetOrRegisterUGauge() {
	m := "server.bytes_sent"
	g := GetOrRegisterUGauge(m, NewRegistry())
	g.Update(47)
	fmt.Println(g.Value()) // Output: 47
}
