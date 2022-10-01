package graphite

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/msaf1980/go-metrics"
)

func floatEquals(a, b float64) bool {
	return (a-b) < 0.000001 && (b-a) < 0.000001
}

func ExampleGraphite() {
	g := New(1*time.Second, "some.prefix", "127.0.0.1:2003")
	g.Start(metrics.DefaultRegistry)
	g.Stop()
}

func ExampleWithConfig() {
	g := WithConfig(&Config{
		Host:          "127.0.0.1:2003",
		FlushInterval: 1 * time.Second,
		DurationUnit:  time.Millisecond,
		Percentiles:   []float64{0.5, 0.75, 0.99, 0.999},
	})
	g.Start(metrics.DefaultRegistry)
	g.Stop()
}

func newTestServer(t *testing.T, prefix string) (map[string]float64, net.Listener, metrics.Registry, *Config, *sync.WaitGroup) {
	res := make(map[string]float64)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("could not start dummy server:", err)
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
				t.Error("dummy server error", err)
			}
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			for err == nil {
				parts := strings.Split(line, " ")
				i, _ := strconv.ParseFloat(parts[1], 64)
				if testing.Verbose() {
					t.Log("recv", parts[0], i, strings.TrimRight(parts[2], "\n"))
				}
				res[parts[0]] = res[parts[0]] + i
				line, err = r.ReadString('\n')
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

	return res, ln, r, c, wg
}

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

func TestWrites(t *testing.T) {
	res, l, r, c, wg := newTestServer(t, "foobar")

	metrics.GetOrRegisterCounter("counter", r).Inc(2)

	metrics.GetOrRegisterGauge("gauge", r).Update(3)
	metrics.GetOrRegisterGaugeFloat64("gauge_float", r).Update(2.1)

	// TODO: Use a mock meter rather than wasting 10s to get a QPS.
	for i := 0; i < 10*4; i++ {
		metrics.GetOrRegisterMeter("meter", r).Mark(1)
		// metrics.GetOrRegisterHistogram("histogram", r, metrics.NewUniformSample(100)).Update(1)
		time.Sleep(200 * time.Millisecond)
	}

	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 5)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 4)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 3)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 2)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 1)

	if err := Once(c, r); err != nil {
		t.Error(err)
	}
	l.Close()
	wg.Wait()

	// counter
	if expected, found := 2.0, res["foobar.counter.count"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.counter.count:", expected, found)
	}
	if expected, found := 200.0, res["foobar.counter.count_ps"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.counter.count:", expected, found)
	}

	// meter
	if expected, found := 40.0, res["foobar.meter.count"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.meter.count:", expected, found)
	}
	if expected, found := 5.0, res["foobar.meter.one-minute"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.meter.one-minute:", expected, found)
	}
	if expected, found := 5.0, res["foobar.meter.five-minute"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.meter.five-minute:", expected, found)
	}
	if expected, found := 5.0, res["foobar.meter.fifteen-minute"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.meter.fifteen-minute:", expected, found)
	}
	if expected, found := 5.12, res["foobar.meter.mean"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.meter.mean:", expected, found)
	}

	// timer [1000, 2000, 3000, 4000, 5000]
	if expected, found := 5.0, res["foobar.timer.count"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.count:", expected, found)
	}
	if expected, found := 500.0, res["foobar.timer.count_ps"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.count_ps:", expected, found)
	}
	if expected, found := 5000.0, res["foobar.timer.999-percentile"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.999-percentile:", expected, found)
	}
	if expected, found := 5000.0, res["foobar.timer.99-percentile"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.99-percentile:", expected, found)
	}
	if expected, found := 4500.0, res["foobar.timer.75-percentile"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.75-percentile:", expected, found)
	}
	if expected, found := 3000.0, res["foobar.timer.50-percentile"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.50-percentile:", expected, found)
	}
	if expected, found := 1000.0, res["foobar.timer.min"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.min:", expected, found)
	}
	if expected, found := 5000.0, res["foobar.timer.max"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.max:", expected, found)
	}
	if expected, found := 3000.0, res["foobar.timer.mean"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.mean:", expected, found)
	}
	if expected, found := 1414.21, res["foobar.timer.std-dev"]; !floatEquals(found, expected) {
		t.Error("bad value foobar.timer.std-dev:", expected, found)
	}
	// 	for psIdx, psKey := range c.Percentiles {
	// 		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
	// 		fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx]/du, now)
	// 	}
	// TODO: may be broken: rate not set
	// if expected, found := 1414.21, res["foobar.timer.one-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.one-minute:", expected, found)
	// }
	// if expected, found := 1414.21, res["foobar.timer.five-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.five-minute:", expected, found)
	// }
	// if expected, found := 1414.21, res["foobar.timer.fifteen-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.fifteen-minute:", expected, found)
	// }

	// histogram
	// case metrics.Gauge:
	// 	fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
	// case metrics.GaugeFloat64:
	// 	fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
	// case metrics.Histogram:
	// 	fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
	// 	fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
	// 	fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
	// 	fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
	// 	fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
	// 	for psIdx, psKey := range c.Percentiles {
	// 		key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
	// 		fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
	// 	}
	// case metrics.Meter:
	// 	m := metric.Snapshot()
	// 	fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
	// 	fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
	// 	fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
	// 	fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
	// 	fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
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

func BenchmarkHistogram(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.GetOrRegisterHistogram("baz", r, s)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Update(int64(i))
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkMeter(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	m := metrics.GetOrRegisterMeter("quux", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Mark(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkTimer(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	t := metrics.GetOrRegisterTimer("bang", r)
	t.Time(func() {})

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Update(47)
		if err := graphite.send(r); err != nil {
			b.Fatal(err)
		}
	}
	graphite.Close()
	b.StopTimer()
	l.Close()
	wg.Wait()
}

func BenchmarkAll(b *testing.B) {
	_, l, r, cfg, wg := newBenchServer(b, "foobar")

	c := metrics.GetOrRegisterCounter("foo", r)

	g := metrics.GetOrRegisterGauge("bar", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64("barf", r)
	gf.Update(47)

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.GetOrRegisterHistogram("baz", r, s)

	t := metrics.GetOrRegisterTimer("bang", r)
	t.Time(func() {})

	m := metrics.GetOrRegisterMeter("quux", r)

	graphite := WithConfig(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		g.Update(1)
		gf.Update(1.1)
		h.Update(int64(i))
		t.Update(47)
		m.Mark(47)
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

	g := metrics.GetOrRegisterGauge("bar", r)
	g.Update(47)

	gf := metrics.GetOrRegisterGaugeFloat64("barf", r)
	gf.Update(47)

	s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
	h := metrics.GetOrRegisterHistogram("baz", r, s)

	t := metrics.GetOrRegisterTimer("bang", r)
	t.Time(func() {})

	m := metrics.GetOrRegisterMeter("quux", r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc(1)
		g.Update(1)
		gf.Update(1.1)
		h.Update(int64(i))
		t.Update(47)
		m.Mark(47)
		if err := Once(cfg, r); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	l.Close()
	wg.Wait()
}
