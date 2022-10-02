package graphite

import (
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/msaf1980/go-metrics"
)

func newBenchServer(b *testing.B, prefix string) (*int64, net.Listener, metrics.Registry, *Config, *sync.WaitGroup) {
	count := new(int64)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal("could not start dummy server:", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var conn net.Conn
			conn, err = ln.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				b.Error("dummy server error", err)
			}
			buf := make([]byte, 64*1024)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					break
				}
				atomic.AddInt64(count, int64(n))
			}
			conn.Close()
		}
	}()

	r := metrics.NewRegistry()

	c := &Config{
		Host:           ln.Addr().String(),
		FlushInterval:  10 * time.Millisecond,
		DurationUnit:   time.Millisecond,
		ConnectTimeout: time.Second,
		Timeout:        time.Second,
		Percentiles:    []float64{0.5, 0.75, 0.99, 0.999},
		Prefix:         prefix,
	}

	return count, ln, r, c, wg
}

func BenchmarkCounter(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounter("foo", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkCounterT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounterT("foo", "tag1=value1;tag21=value21", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkGauge(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	g := metrics.GetOrRegisterGauge("bar", r)
	g.Update(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkGaugeT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	g := metrics.GetOrRegisterGaugeT("bar", "tag1=value1;tag21=value21", r)
	g.Update(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkGaugeFloat64(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	g := metrics.GetOrRegisterGaugeFloat64("bar", r)
	g.Update(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(1.1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkGaugeFloat64T(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	g := metrics.GetOrRegisterGaugeFloat64T("bar", "tag1=value1;tag21=value21", r)
	g.Update(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Update(1.1)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

// func BenchmarkHistogram(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
// 	h := metrics.GetOrRegisterHistogram("baz", r, s)

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		h.Update(int64(i))
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

// func BenchmarkHistogramT(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
// 	h := metrics.GetOrRegisterHistogramT("baz", "tag1=value1;tag21=value21", r, s)

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		h.Update(int64(i))
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

// func BenchmarkMeter(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	m := metrics.GetOrRegisterMeter("quux", r)

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		m.Mark(47)
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

// func BenchmarkMeterT(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	m := metrics.GetOrRegisterMeterT("quux", "tag1=value1;tag21=value21", r)

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		m.Mark(47)
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

// func BenchmarkTimer(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	t := metrics.GetOrRegisterTimer("bang", r)
// 	t.Time(func() {})

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		t.Update(47)
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

// func BenchmarkTimerT(b *testing.B) {
// 	_, l, r, cfg, wg := newBenchServer(b, "foobar")

// 	t := metrics.GetOrRegisterTimerT("bang", "tag1=value1;tag21=value21", r)
// 	t.Time(func() {})

// 	graphite := WithConfig(cfg)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		t.Update(47)
// 		if err := graphite.send(r); err != nil {
// 			b.Fatal(err)
// 		}
// 	}
// 	graphite.Close()
// 	b.StopTimer()
// 	l.Close()
// 	wg.Wait()
// }

func BenchmarkAll(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounter("foo", r)

	differ := metrics.GetOrRegisterDiffer("differ", r)
	differ.Update(1)

	g := metrics.GetOrRegisterGauge("bar", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64("barf", r)
	gf.Update(47)

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogram("baz", r, s)

	// t := metrics.GetOrRegisterTimer("bang", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeter("quux", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		differ.Update(int64(i))
		g.Update(1)
		gf.Update(1.1)
		// h.Update(int64(i))
		// t.Update(47)
		// m.Mark(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkAllT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounterT("foo", "tag1=value1;tag21=value21", r)

	differ := metrics.GetOrRegisterDifferT("differ", "tag1=value1;tag21=value21", r)
	differ.Update(1)

	g := metrics.GetOrRegisterGaugeT("bar", "tag1=value1;tag21=value21", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64T("barf", "tag1=value1;tag21=value21", r)
	gf.Update(47)

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogramT("baz", "tag1=value1;tag21=value21", r, s)

	// t := metrics.GetOrRegisterTimerT("bang", "tag1=value1;tag21=value21", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeterT("quux", "tag1=value1;tag21=value21", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		differ.Update(int64(i))
		g.Update(1)
		gf.Update(1.1)
		// h.Update(int64(i))
		// t.Update(47)
		// m.Mark(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkOnce(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounter("foo", r)

	differ := metrics.GetOrRegisterDiffer("differ", r)
	differ.Update(1)

	g := metrics.GetOrRegisterGauge("bar", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64("barf", r)
	gf.Update(47)

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogram("baz", r, s)

	// t := metrics.GetOrRegisterTimer("bang", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeter("quux", r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		differ.Update(int64(i))
		g.Update(1)
		gf.Update(1.1)
		// h.Update(int64(i))
		// t.Update(47)
		// m.Mark(47)
		if err := Once(cfg, r); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkOnceT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounterT("foo", "tag1=value1;tag21=value21", r)

	differ := metrics.GetOrRegisterDifferT("differ", "tag1=value1;tag21=value21", r)
	differ.Update(1)

	g := metrics.GetOrRegisterGaugeT("bar", "tag1=value1;tag21=value21", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64T("barf", "tag1=value1;tag21=value21", r)
	gf.Update(47)

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogramT("baz", "tag1=value1;tag21=value21", r, s)

	// t := metrics.GetOrRegisterTimerT("bang", "tag1=value1;tag21=value21", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeterT("quux", "tag1=value1;tag21=value21", r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		differ.Update(int64(i))
		g.Update(1)
		gf.Update(1.1)
		// h.Update(int64(i))
		// t.Update(47)
		// m.Mark(47)
		if err := Once(cfg, r); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	l.Close()
	wg.Wait()
}
