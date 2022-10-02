package metrics

import "testing"

func BenchmarkDiffer(b *testing.B) {
	g := NewDiffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(int64(i))
	}
}
