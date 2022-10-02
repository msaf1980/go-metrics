package graphite

import (
	"testing"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-metrics/test"
)

func TestWritesT(t *testing.T) {
	length := 0
	res, l, r, c, wg := newTestServer(t, "foobar")
	defer r.UnregisterAll()

	metrics.GetOrRegisterCounterT("counter", ";tag1=value1;tag21=value21", r).Inc(2)
	length += 2

	metrics.GetOrRegisterDifferT("differ", ";tag1=value1;tag21=value21", r).Update(3)
	metrics.GetOrRegisterDifferT("differ", ";tag1=value1;tag21=value21", r).Update(9)
	metrics.GetOrRegisterGauge("gauge", r).Update(4) // non tagged
	metrics.GetOrRegisterGaugeT("gauge", ";tag1=value1;tag21=value21", r).Update(3)
	metrics.GetOrRegisterGaugeFloat64T("gauge_float", ";tag1=value1;tag21=value21", r).Update(2.1)
	length += 3

	// metrics.GetOrRegisterTimerT("timer", ";tag1=value1;tag21=value21", r).Update(time.Second * 5)
	// metrics.GetOrRegisterTimerT("timer", ";tag1=value1;tag21=value21", r).Update(time.Second * 4)
	// metrics.GetOrRegisterTimerT("timer", ";tag1=value1;tag21=value21", r).Update(time.Second * 3)
	// metrics.GetOrRegisterTimerT("timer", ";tag1=value1;tag21=value21", r).Update(time.Second * 2)
	// metrics.GetOrRegisterTimerT("timer", ";tag1=value1;tag21=value21", r).Update(time.Second * 1)
	// length += 10

	// // TODO: Use a mock meter rather than wasting 10s to get a QPS.
	// for i := 0; i < 10*4; i++ {
	// 	metrics.GetOrRegisterMeterT("meter", ";tag1=value1;tag21=value21", r).Mark(1)
	// 	// metrics.GetOrRegisterHistogram("histogram", r, metrics.NewUniformSample(100)).Update(1)
	// 	time.Sleep(200 * time.Millisecond)
	// }
	// length += 5 + 4

	if err := Once(c, r); err != nil {
		t.Error(err)
	}
	l.Close()
	wg.Wait()

	// if len(res) != length {
	// 	for k, v := range res {
	// 		fmt.Printf(" %s = %#v\n", k, v)
	// 	}
	// 	t.Errorf("want %d metrics, got %d", length, len(res))
	// }

	want := map[string]test.Value{
		// count
		"foobar.counter.count;tag1=value1;tag21=value21":    {V: 2.0},
		"foobar.counter.count_ps;tag1=value1;tag21=value21": {V: 200.0},
		// gauge
		"foobar.differ;tag1=value1;tag21=value21": {V: 6.0},
		"foobar.gauge":                                 {V: 4.0},
		"foobar.gauge;tag1=value1;tag21=value21":       {V: 3.0},
		"foobar.gauge_float;tag1=value1;tag21=value21": {V: 2.1},
		// // meter
		// "foobar.meter.count;tag1=value1;tag21=value21":          {V: 40.0},
		// "foobar.meter.mean;tag1=value1;tag21=value21":           {V: 5.12},
		// "foobar.meter.one-minute;tag1=value1;tag21=value21":     {V: 5.0},
		// "foobar.meter.five-minute;tag1=value1;tag21=value21":    {V: 5.0},
		// "foobar.meter.fifteen-minute;tag1=value1;tag21=value21": {V: 5.0},
		// // timer
		// "foobar.timer.count;tag1=value1;tag21=value21":          {V: 5.0},
		// "foobar.timer.count_ps;tag1=value1;tag21=value21":       {V: 500.0},
		// "foobar.timer.min;tag1=value1;tag21=value21":            {V: 1000.0},
		// "foobar.timer.max;tag1=value1;tag21=value21":            {V: 5000.0},
		// "foobar.timer.std-dev;tag1=value1;tag21=value21":        {V: 1414.21},
		// "foobar.timer.mean;tag1=value1;tag21=value21":           {V: 3000.0},
		// "foobar.timer.mean-rate;tag1=value1;tag21=value21":      {V: 240419.29, Dev: 250000},
		// "foobar.timer.50-percentile;tag1=value1;tag21=value21":  {V: 3000.0},
		// "foobar.timer.75-percentile;tag1=value1;tag21=value21":  {V: 4500.0},
		// "foobar.timer.99-percentile;tag1=value1;tag21=value21":  {V: 5000.0},
		// "foobar.timer.999-percentile;tag1=value1;tag21=value21": {V: 5000.0},
		// "foobar.timer.fifteen-minute;tag1=value1;tag21=value21": {V: 1.0, Dev: 0.2},
		// "foobar.timer.five-minute;tag1=value1;tag21=value21":    {V: 1.0, Dev: 0.2},
		// "foobar.timer.one-minute;tag1=value1;tag21=value21":     {V: 1.0, Dev: 0.2},
	}

	test.CompareMetrics(t, want, res)

	// // counter
	// if expected, found := 2.0, res["foobar.counter.count;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.counter.count;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 200.0, res["foobar.counter.count_ps;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.counter.count;tag1=value1;tag21=value21:", expected, found)
	// }

	// // gauge
	// if expected, found := 4.0, res["foobar.gauge"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.gauge;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 3.0, res["foobar.gauge;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.gauge;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 2.1, res["foobar.gauge_float;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.gauge_float;tag1=value1;tag21=value21:", expected, found)
	// }

	// meter
	// if expected, found := 40.0, res["foobar.meter.count;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.count;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.one-minute;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.one-minute;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.five-minute;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.five-minute;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5.0, res["foobar.meter.fifteen-minute;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.fifteen-minute;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5.12, res["foobar.meter.mean;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.meter.mean;tag1=value1;tag21=value21:", expected, found)
	// }

	// timer [1000, 2000, 3000, 4000, 5000]
	// if expected, found := 5.0, res["foobar.timer.count;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.count;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 500.0, res["foobar.timer.count_ps;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.count_p;tag1=value1;tag21=value21s:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.999-percentile;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.999-percentile;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.99-percentile;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.99-percentile;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 4500.0, res["foobar.timer.75-percentile;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.75-percentile;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 3000.0, res["foobar.timer.50-percentile;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.50-percentile;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 1000.0, res["foobar.timer.min;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.min;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 5000.0, res["foobar.timer.max;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.max;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 3000.0, res["foobar.timer.mean;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.mean;tag1=value1;tag21=value21:", expected, found)
	// }
	// if expected, found := 1414.21, res["foobar.timer.std-dev;tag1=value1;tag21=value21"]; !floatEquals(found, expected) {
	// 	t.Error("bad value foobar.timer.std-dev;tag1=value1;tag21=value21:", expected, found)
	// }
}
