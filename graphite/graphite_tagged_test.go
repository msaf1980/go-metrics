package graphite

import (
	"testing"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-metrics/test"
)

func TestWritesT(t *testing.T) {
	sb, res, l, r, c, wg := newTestServer(t, "foobar", "footag")
	defer r.UnregisterAll()

	metrics.GetOrRegisterCounterT("counter", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Inc(2)
	// check for no conflicts with tagged counter
	metrics.GetOrRegisterCounter("counter", r).Inc(2)

	metrics.GetOrRegisterDownCounterT("dcounter", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Dec(4)

	metrics.GetOrRegisterDifferT("differ", map[string]string{"tag1": "value1", "tag21": "value21"}, r, 0).Update(3)
	metrics.GetOrRegisterDifferT("differ", map[string]string{"tag1": "value1", "tag21": "value21"}, r, 0).Update(9)
	metrics.GetOrRegisterGauge("gauge", r).Update(4) // non tagged
	metrics.GetOrRegisterGaugeT("gauge", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(3)
	metrics.GetOrRegisterGaugeFloat64T("gauge_float", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(2.1)

	h := metrics.NewVHistogram([]int64{1, 2, 5, 8, 20}, nil)
	h.Add(2)
	h.Add(6)
	if err := r.RegisterT("histogram", map[string]string{"tag1": "value1", "tag21": "value21"}, h); err != nil {
		l.Close()
		wg.Wait()
		t.Fatal(err)
	}

	sh := metrics.GetOrRegisterFixedSumHistogramT("shistogram", map[string]string{"tag1": "value1", "tag21": "value21"}, r, 1, 3, 1).AddLabelPrefix("req_")
	sh.Add(2)
	sh.Add(6)

	rate := metrics.GetOrRegisterRateT("ratefoo", map[string]string{"tag1": "value1", "tag21": "value21"}, r).SetName("_value").SetRateName("_rate")
	rate.UpdateTs(1, 1e9)
	rate.UpdateTs(7, 3e9)

	rate2 := metrics.GetOrRegisterRateT("ratefoo2", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	rate2.UpdateTs(2, 1e9)
	rate2.UpdateTs(8, 3e9)

	// metrics.GetOrRegisterTimerT("timer", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(time.Second * 5)
	// metrics.GetOrRegisterTimerT("timer", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(time.Second * 4)
	// metrics.GetOrRegisterTimerT("timer", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(time.Second * 3)
	// metrics.GetOrRegisterTimerT("timer", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(time.Second * 2)
	// metrics.GetOrRegisterTimerT("timer", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Update(time.Second * 1)
	// length += 10

	// // TODO: Use a mock meter rather than wasting 10s to get a QPS.
	// for i := 0; i < 10*4; i++ {
	// 	metrics.GetOrRegisterMeterT("meter", map[string]string{"tag1": "value1", "tag21": "value21"}, r).Mark(1)
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
		"footag.counter;tag1=value1;tag21=value21": {V: 2.0},
		"foobar.counter": {V: 2.0},
		"footag.dcounter;tag1=value1;tag21=value21": {V: -4.0},
		// gauge
		"footag.differ;tag1=value1;tag21=value21": {V: 6.0},
		"foobar.gauge":                                 {V: 4.0},
		"footag.gauge;tag1=value1;tag21=value21":       {V: 3.0},
		"footag.gauge_float;tag1=value1;tag21=value21": {V: 2.1},
		// histogram
		"footag.histogram.1;tag1=value1;tag21=value21;le=1":     {V: 0},
		"footag.histogram.2;tag1=value1;tag21=value21;le=2":     {V: 1},
		"footag.histogram.5;tag1=value1;tag21=value21;le=5":     {V: 0},
		"footag.histogram.8;tag1=value1;tag21=value21;le=8":     {V: 1},
		"footag.histogram.20;tag1=value1;tag21=value21;le=20":   {V: 0},
		"footag.histogram.inf;tag1=value1;tag21=value21;le=inf": {V: 0},
		"footag.histogram.total;tag1=value1;tag21=value21":      {V: 2},
		// shistogram
		"footag.shistogram.req_1;tag1=value1;tag21=value21;le=1":     {V: 2.0},
		"footag.shistogram.req_2;tag1=value1;tag21=value21;le=2":     {V: 2.0},
		"footag.shistogram.req_3;tag1=value1;tag21=value21;le=3":     {V: 1.0},
		"footag.shistogram.req_inf;tag1=value1;tag21=value21;le=inf": {V: 1.0},
		"footag.shistogram.total;tag1=value1;tag21=value21":          {V: 2.0},
		// rate
		"footag.ratefoo_value;tag1=value1;tag21=value21":  {V: 7},
		"footag.ratefoo_rate;tag1=value1;tag21=value21":   {V: 3},
		"footag.ratefoo2.value;tag1=value1;tag21=value21": {V: 8},
		"footag.ratefoo2.rate;tag1=value1;tag21=value21":  {V: 3},
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

	if test.CompareMetrics(t, want, res) {
		t.Errorf("dump received\n%s", sb.String())
	}
}
