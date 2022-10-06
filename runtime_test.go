package metrics

import (
	"runtime"
	"testing"
	"time"
)

func TestRuntimeMemStatsDoubleRegister(t *testing.T) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	storedGauge := r.Get(RuntimeNames.MemStats.LastGC).(Gauge)

	runtime.GC()
	CaptureRuntimeMemStatsOnce()

	firstGC := storedGauge.Value()
	if firstGC == 0 {
		t.Errorf("firstGC got %d, expected timestamp > 0", firstGC)
	}

	time.Sleep(time.Millisecond)

	RegisterRuntimeMemStats(r)
	runtime.GC()
	CaptureRuntimeMemStatsOnce()
	if lastGC := storedGauge.Value(); firstGC == lastGC {
		t.Errorf("lastGC got %d, expected a higher timestamp value", lastGC)
	}
}

func BenchmarkRuntimeMemStats(b *testing.B) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CaptureRuntimeMemStatsOnce()
	}
}

// func TestRuntimeMemStats(t *testing.T) {
// 	r := NewRegistry()
// 	RegisterRuntimeMemStats(r)
// 	CaptureRuntimeMemStatsOnce()
// 	zero := runtimeMetrics.MemStats.PauseTotalNs.Value() // Get a "zero" since GC may have run before these tests.
// 	runtime.GC()
// 	CaptureRuntimeMemStatsOnce()
// 	if count := runtimeMetrics.MemStats.PauseNs.Count(); count-zero != 1 {
// 		t.Fatal(count - zero)
// 	}
// 	runtime.GC()
// 	runtime.GC()
// 	CaptureRuntimeMemStatsOnce()
// 	if count := runtimeMetrics.MemStats.PauseNs.Count(); count-zero != 3 {
// 		t.Fatal(count - zero)
// 	}
// 	for i := 0; i < 256; i++ {
// 		runtime.GC()
// 	}
// 	CaptureRuntimeMemStatsOnce()
// 	if count := runtimeMetrics.MemStats.PauseNs.Count(); count-zero != 259 {
// 		t.Fatal(count - zero)
// 	}
// 	for i := 0; i < 257; i++ {
// 		runtime.GC()
// 	}
// 	CaptureRuntimeMemStatsOnce()
// 	if count := runtimeMetrics.MemStats.PauseNs.Count(); count-zero != 515 { // We lost one because there were too many GCs between captures.
// 		t.Fatal(count - zero)
// 	}
// }

func TestRuntimeMemStatsNumThread(t *testing.T) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)
	CaptureRuntimeMemStatsOnce()

	if value := runtimeMetrics.NumThread.Value(); value < 1 {
		t.Fatalf("got NumThread: %d, wanted at least 1", value)
	}
}

func TestRuntimeMemStatsBlocking(t *testing.T) {
	if g := runtime.GOMAXPROCS(0); g < 2 {
		t.Skipf("skipping TestRuntimeMemStatsBlocking with GOMAXPROCS=%d\n", g)
	}
	ch := make(chan int)
	go testRuntimeMemStatsBlocking(ch)
	var memStats runtime.MemStats
	t0 := time.Now()
	runtime.ReadMemStats(&memStats)
	t1 := time.Now()
	t.Log("i++ during runtime.ReadMemStats:", <-ch)
	go testRuntimeMemStatsBlocking(ch)
	d := t1.Sub(t0)
	t.Log(d)
	time.Sleep(d)
	t.Log("i++ during time.Sleep:", <-ch)
}

func testRuntimeMemStatsBlocking(ch chan int) {
	i := 0
	for {
		select {
		case ch <- i:
			return
		default:
			i++
		}
	}
}

func TestRuntimeStats(t *testing.T) {
	r := NewRegistry()
	RegisterRuntimeMemStats(r)

	CaptureRuntimeMemStatsOnce()
	b := make([]byte, 1024)
	_ = b
	go func() {}()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	CaptureRuntimeMemStatsOnce()

	alloc1 := runtimeMetrics.MemStats.Alloc.Value()
	backhashsys1 := runtimeMetrics.MemStats.BuckHashSys.Value()
	frees1, frees_rate1 := runtimeMetrics.MemStats.Frees.Values()
	malloc1, malloc_rate1 := runtimeMetrics.MemStats.Mallocs.Values()
	gc1, gc_rate1 := runtimeMetrics.MemStats.NumGC.Values()
	cgo1, cgo_rate1 := runtimeMetrics.NumCgoCall.Values()
	threads1 := runtimeMetrics.NumThread.Value()
	goroutines1 := runtimeMetrics.NumGoroutine.Value()

	{
		d := make([]byte, 1024)
		_ = d
		b = nil
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
	}()
	time.Sleep(100 * time.Millisecond)

	runtime.GC()
	CaptureRuntimeMemStatsOnce()

	alloc2 := runtimeMetrics.MemStats.Alloc.Value()
	backhashsys2 := runtimeMetrics.MemStats.BuckHashSys.Value()
	frees2, frees_rate2 := runtimeMetrics.MemStats.Frees.Values()
	malloc2, malloc_rate2 := runtimeMetrics.MemStats.Mallocs.Values()
	gc2, gc_rate2 := runtimeMetrics.MemStats.NumGC.Values()
	cgo2, cgo_rate2 := runtimeMetrics.NumCgoCall.Values()
	threads2 := runtimeMetrics.NumThread.Value()
	goroutines2 := runtimeMetrics.NumGoroutine.Value()

	t.Logf("%s: #1 %d", RuntimeNames.MemStats.Alloc, alloc1)
	t.Logf("%s: #2 %d", RuntimeNames.MemStats.Alloc, alloc2)
	t.Logf("%s: #1 %d", RuntimeNames.MemStats.BuckHashSys, backhashsys1)
	t.Logf("%s: #2 %d", RuntimeNames.MemStats.BuckHashSys, backhashsys2)
	t.Logf("%s: #1 %d %f/s", RuntimeNames.MemStats.Frees, frees1, frees_rate1)
	t.Logf("%s: #2 %d %f/s", RuntimeNames.MemStats.Frees, frees2, frees_rate2)
	t.Logf("%s: #1 %d %f/s", RuntimeNames.MemStats.Mallocs, malloc1, malloc_rate1)
	t.Logf("%s: #2 %d %f/s", RuntimeNames.MemStats.Mallocs, malloc2, malloc_rate2)
	t.Logf("%s: #1 %d %f/s", RuntimeNames.MemStats.NumGC, gc1, gc_rate1)
	t.Logf("%s: #2 %d %f/s", RuntimeNames.MemStats.NumGC, gc2, gc_rate2)
	t.Logf("%s: #1 %d %f/s", RuntimeNames.NumCgoCall, cgo1, cgo_rate1)
	t.Logf("%s: #2 %d %f/s", RuntimeNames.NumCgoCall, cgo2, cgo_rate2)
	t.Logf("%s: #1 %d", RuntimeNames.NumThread, threads1)
	t.Logf("%s: #2 %d", RuntimeNames.NumThread, threads2)
	t.Logf("%s: #1 %d", RuntimeNames.NumGoroutine, goroutines1)
	t.Logf("%s: #2 %d", RuntimeNames.NumGoroutine, goroutines2)

	if goroutines2 != goroutines1+1 {
		t.Error("goroutines", goroutines1, goroutines2)
	}
}
