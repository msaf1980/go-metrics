package graphite

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-metrics/test"
)

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

func TestWrites(t *testing.T) {
	length := 0
	res, l, r, c, wg := newTestServer(t, "foobar")
	defer r.UnregisterAll()

	metrics.GetOrRegisterCounter("counter", r).Inc(2)
	length += 2

	metrics.GetOrRegisterGauge("gauge", r).Update(3)
	metrics.GetOrRegisterGaugeFloat64("gauge_float", r).Update(2.1)
	length += 2

	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 5)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 4)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 3)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 2)
	metrics.GetOrRegisterTimer("timer", r).Update(time.Second * 1)
	length += 10

	// TODO: Use a mock meter rather than wasting 10s to get a QPS.
	for i := 0; i < 10*4; i++ {
		metrics.GetOrRegisterMeter("meter", r).Mark(1)
		// metrics.GetOrRegisterHistogram("histogram", r, metrics.NewUniformSample(100)).Update(1)
		time.Sleep(200 * time.Millisecond)
	}
	length += 5 + 4

	if err := Once(c, r); err != nil {
		t.Error(err)
	}
	l.Close()
	wg.Wait()

	if len(res) != length {
		for k, v := range res {
			fmt.Printf(" %s = %#v\n", k, v)
		}
		t.Errorf("want %d metrics, got %d", length, len(res))
	}

	want := map[string]test.Value{
		// count
		"foobar.counter.count":    {V: 2.0},
		"foobar.counter.count_ps": {V: 200.0},
		// gauge
		"foobar.gauge":       {V: 3.0},
		"foobar.gauge_float": {V: 2.1},
		// meter
		"foobar.meter.count":          {V: 40.0},
		"foobar.meter.mean":           {V: 5.12},
		"foobar.meter.one-minute":     {V: 5.0},
		"foobar.meter.five-minute":    {V: 5.0},
		"foobar.meter.fifteen-minute": {V: 5.0},
		// timer
		"foobar.timer.count":          {V: 5.0},
		"foobar.timer.count_ps":       {V: 500.0},
		"foobar.timer.min":            {V: 1000.0},
		"foobar.timer.max":            {V: 5000.0},
		"foobar.timer.std-dev":        {V: 1414.21},
		"foobar.timer.mean":           {V: 3000.0},
		"foobar.timer.mean-rate":      {V: 240419.29, Dev: 250000},
		"foobar.timer.50-percentile":  {V: 3000.0},
		"foobar.timer.75-percentile":  {V: 4500.0},
		"foobar.timer.99-percentile":  {V: 5000.0},
		"foobar.timer.999-percentile": {V: 5000.0},
		"foobar.timer.fifteen-minute": {V: 1.0, Dev: 0.2},
		"foobar.timer.five-minute":    {V: 1.0, Dev: 0.2},
		"foobar.timer.one-minute":     {V: 1.0, Dev: 0.2},
	}

	test.CompareMetrics(t, want, res)

	// // counter
	// if expected, found := 2.0, res["foobar.counter.count"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.counter.count:", expected, found)
	// }
	// if expected, found := 200.0, res["foobar.counter.count_ps"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.counter.count:", expected, found)
	// }

	// // gauge
	// if expected, found := 3.0, res["foobar.gauge"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.gauge:", expected, found)
	// }
	// if expected, found := 2.1, res["foobar.gauge_float"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.gauge_float:", expected, found)
	// }

	// // meter
	// if expected, found := 40.0, res["foobar.meter.count"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.count:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.one-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.one-minute:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.five-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.five-minute:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.fifteen-minute"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.fifteen-minute:", expected, found)
	// }
	// if expected, found := 5.12, res["foobar.meter.mean"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.mean:", expected, found)
	// }

	// // timer [1000, 2000, 3000, 4000, 5000]
	// if expected, found := 5.0, res["foobar.timer.count"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.count:", expected, found)
	// }
	// if expected, found := 500.0, res["foobar.timer.count_ps"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.count_ps:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.999-percentile"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.999-percentile:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.99-percentile"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.99-percentile:", expected, found)
	// }
	// if expected, found := 4500.0, res["foobar.timer.75-percentile"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.75-percentile:", expected, found)
	// }
	// if expected, found := 3000.0, res["foobar.timer.50-percentile"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.50-percentile:", expected, found)
	// }
	// if expected, found := 1000.0, res["foobar.timer.min"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.min:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.max"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.max:", expected, found)
	// }
	// if expected, found := 3000.0, res["foobar.timer.mean"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.mean:", expected, found)
	// }
	// if expected, found := 1414.21, res["foobar.timer.std-dev"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.std-dev:", expected, found)
	// }
}
