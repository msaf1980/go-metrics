package metrics

import "testing"

func BenchmarkDiffer(b *testing.B) {
	g := NewDiffer(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(int64(i))
	}
}
