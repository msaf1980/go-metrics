package metrics

import (
	"fmt"
	"testing"
)

func BenchmarkDiffer(b *testing.B) {
	g := NewDiffer(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(int64(i))
	}
}

func TestDiffer(t *testing.T) {
	g := NewDiffer(1)
	g.Update(47)
	if v := g.Value(); v != 46 {
		t.Errorf("g.Value(): 46 != %v\n", v)
	}
}

func TestDifferSnapshot(t *testing.T) {
	g := NewDiffer(0)
	g.Update(47)
	snapshot := g.Snapshot()
	g.Update(48)
	if v := snapshot.Value(); v != 47 {
		t.Errorf("g.Value(): 47 != %v\n", v)
	}
}

func TestGetOrRegisterDiffer(t *testing.T) {
	r := NewRegistry()
	NewRegisteredDiffer("foo", r, 2).Update(47)
	if g := GetOrRegisterDiffer("foo", r, 1); g.Value() != 45 {
		t.Fatal(g)
	}
}

func ExampleGetOrRegisterDiffer() {
	m := "server.memory_used"
	init := int64(1)
	g := GetOrRegisterDiffer(m, NewRegistry(), init)
	g.Update(47)
	fmt.Println(g.Value()) // Output: 46
}
