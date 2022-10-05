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

	c := metrics.GetOrRegisterCounterT("foo", map[string]string{"tag1": "value1", "tag21": "value21"}, r)

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

	g := metrics.GetOrRegisterGaugeT("bar", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
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

	g := metrics.GetOrRegisterGaugeFloat64T("bar", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
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

func BenchmarkVHistogram(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	h := metrics.NewVHistogram([]int64{10, 50, 100, 200, 300}, nil, "", "")
	if err := r.Register("histogram", h); err != nil {
		l.Close()
		wg.Wait()
		b.Fatal(err)
	}
	h.Add(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkVHistogramT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	h := metrics.NewVHistogram([]int64{10, 50, 100, 200, 300}, nil, "", "")
	if err := r.RegisterT("histogram", map[string]string{"tag1": "value1", "tag21": "value21"}, h); err != nil {
		l.Close()
		wg.Wait()
		b.Fatal(err)
	}
	h.Add(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkFVHistogram(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	h := metrics.NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil, "", "")
	if err := r.Register("histogram", h); err != nil {
		l.Close()
		wg.Wait()
		b.Fatal(err)
	}
	h.Add(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkFVHistogramT(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	h := metrics.NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil, "", "")
	if err := r.RegisterT("histogram", map[string]string{"tag1": "value1", "tag21": "value21"}, h); err != nil {
		l.Close()
		wg.Wait()
		b.Fatal(err)
	}
	h.Add(47)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(47)
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
// 	h := metrics.GetOrRegisterHistogramT("baz", map[string]string{"tag1": "value1", "tag21": "value21"}, r, s)

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

// 	m := metrics.GetOrRegisterMeterT("quux", map[string]string{"tag1": "value1", "tag21": "value21"}, r)

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

// 	t := metrics.GetOrRegisterTimerT("bang", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
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

	// 100 counters
	c := make([]metrics.Counter, 100)
	for i := 0; i < len(c); i++ {
		c[i] = metrics.GetOrRegisterCounter("foo", r)
	}

	// 100 number gauges
	differ := make([]metrics.Gauge, 50)
	for i := 0; i < len(differ); i++ {
		differ[i] = metrics.GetOrRegisterDiffer("differ", r)
	}
	g := make([]metrics.Gauge, 50)
	for i := 0; i < len(g); i++ {
		g[i] = metrics.GetOrRegisterGauge("bar", r)
	}

	// 100 float gauges
	gf := make([]metrics.GaugeFloat64, 50)
	for i := 0; i < len(gf); i++ {
		gf[i] = metrics.GetOrRegisterGaugeFloat64("barf", r)
	}

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogram("baz", r, s)

	// t := metrics.GetOrRegisterTimer("bang", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeter("quux", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(c); i++ {
			c[i].Inc(1)
		}
		for i := 0; i < len(differ); i++ {
			differ[i].Update(int64(i))
		}
		for i := 0; i < len(g); i++ {
			g[i].Update(1)
		}
		for i := 0; i < len(gf); i++ {
			gf[i].Update(1.1)
		}
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

	// 100 counters and ucounts
	c := make([]metrics.Counter, 100)
	for i := 0; i < len(c); i++ {
		c[i] = metrics.GetOrRegisterCounterT("foo", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	}

	// 100 number gauges
	differ := make([]metrics.Gauge, 50)
	for i := 0; i < len(differ); i++ {
		differ[i] = metrics.GetOrRegisterDifferT("differ", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	}
	g := make([]metrics.Gauge, 50)
	for i := 0; i < len(g); i++ {
		g[i] = metrics.GetOrRegisterGaugeT("bar", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	}

	// 100 float gauges
	gf := make([]metrics.GaugeFloat64, 50)
	for i := 0; i < len(gf); i++ {
		gf[i] = metrics.GetOrRegisterGaugeFloat64T("barf", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	}

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogram("baz", r, s)

	// t := metrics.GetOrRegisterTimer("bang", r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeter("quux", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(c); i++ {
			c[i].Inc(1)
		}
		for i := 0; i < len(differ); i++ {
			differ[i].Update(int64(i))
		}
		for i := 0; i < len(g); i++ {
			g[i].Update(1)
		}
		for i := 0; i < len(gf); i++ {
			gf[i].Update(1.1)
		}
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

	c := metrics.GetOrRegisterCounterT("foo", map[string]string{"tag1": "value1", "tag21": "value21"}, r)

	differ := metrics.GetOrRegisterDifferT("differ", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	differ.Update(1)

	g := metrics.GetOrRegisterGaugeT("bar", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64T("barf", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	gf.Update(47)

	// s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	// h := metrics.GetOrRegisterHistogramT("baz", map[string]string{"tag1": "value1", "tag21": "value21"}, r, s)

	// t := metrics.GetOrRegisterTimerT("bang", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	// t.Time(func() {})

	// m := metrics.GetOrRegisterMeterT("quux", map[string]string{"tag1": "value1", "tag21": "value21"}, r)

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
