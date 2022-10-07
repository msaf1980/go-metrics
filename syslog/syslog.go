//go:build !windows
// +build !windows

package metrics

import (
	"fmt"
	"log/syslog"
	"time"

	"github.com/msaf1980/go-metrics"
)

// Output each metric in the given registry to syslog periodically using
// the given syslogger.
func Syslog(r metrics.Registry, d time.Duration, w *syslog.Writer, minLock bool) {
	for range time.Tick(d) {
		r.Each(func(name, tags string, tagsMap map[string]string, i interface{}) error {
			switch metric := i.(type) {
			case metrics.Counter:
				w.Info(fmt.Sprintf("counter %s%s count: %d", name, tags, metric.Count()))
			case metrics.DownCounter:
				w.Info(fmt.Sprintf("counter %s%s count: %d", name, tags, metric.Count()))
			case metrics.Gauge:
				w.Info(fmt.Sprintf("gauge %s%s value: %d", name, tags, metric.Value()))
			case metrics.GaugeFloat64:
				w.Info(fmt.Sprintf("gauge %s%s: value: %f", name, tags, metric.Value()))
			case metrics.Healthcheck:
				w.Info(fmt.Sprintf("healthcheck %s%s up: %d", name, tags, metric.Check()))
			case metrics.HistogramInterface:
				vals := metric.Values()
				if metric.IsSummed() {
					for i, label := range metric.Labels() {
						w.Info(fmt.Sprintf("histogram %s%s %s value: %d", name, tags, label, vals[i]))
					}
					w.Info(fmt.Sprintf("histogram %s%s %s total: %d", name, tags, metric.NameTotal(), vals[0]))
				} else {
					var total uint64
					for i, label := range metric.Labels() {
						w.Info(fmt.Sprintf("histogram %s%s %s value: %d", name, tags, label, vals[i]))
						total += vals[i]
					}
					w.Info(fmt.Sprintf("histogram %s%s %s total: %d", name, tags, metric.NameTotal(), total))
				}
			case metrics.Rate:
				v, rate := metric.Values()
				w.Info(fmt.Sprintf("rate %s%s%s value: %d rate: %f\n", name, metric.Name(), tags, v, rate))
				w.Info(fmt.Sprintf("rate %s%s%s value: %d rate: %f\n", name, metric.RateName(), tags, v, rate))
			case metrics.FRate:
				v, rate := metric.Values()
				w.Info(fmt.Sprintf("rate %s%s%s value: %f rate: %f\n", name, metric.Name(), tags, v, rate))
				w.Info(fmt.Sprintf("rate %s%s%s value: %f rate: %f\n", name, metric.RateName(), tags, v, rate))
				// case metrics.Histogram:
				// 	h := metric.Snapshot()
				// 	ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				// 	w.Info(fmt.Sprintf(
				// 		"histogram %s%s count: %d min: %d max: %d mean: %.2f stddev: %.2f median: %.2f 75%%: %.2f 95%%: %.2f 99%%: %.2f 99.9%%: %.2f",
				// 		name, tags,
				// 		h.Count(),
				// 		h.Min(),
				// 		h.Max(),
				// 		h.Mean(),
				// 		h.StdDev(),
				// 		ps[0],
				// 		ps[1],
				// 		ps[2],
				// 		ps[3],
				// 		ps[4],
				// 	))
				// case metrics.Meter:
				// 	m := metric.Snapshot()
				// 	w.Info(fmt.Sprintf(
				// 		"meter %s%s count: %d 1-min: %.2f 5-min: %.2f 15-min: %.2f mean: %.2f",
				// 		name, tags,
				// 		m.Count(),
				// 		m.Rate1(),
				// 		m.Rate5(),
				// 		m.Rate15(),
				// 		m.RateMean(),
				// 	))
				// case metrics.Timer:
				// 	t := metric.Snapshot()
				// 	ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				// 	w.Info(fmt.Sprintf(
				// 		"timer %s%s count: %d min: %d max: %d mean: %.2f stddev: %.2f median: %.2f 75%%: %.2f 95%%: %.2f 99%%: %.2f 99.9%%: %.2f mean-rate: %.2f\n",
				// 		// " 1-min: %.2f 5-min: %.2f 15-min: %.2f",
				// 		name, tags,
				// 		t.Count(),
				// 		t.Min(),
				// 		t.Max(),
				// 		t.Mean(),
				// 		t.StdDev(),
				// 		ps[0],
				// 		ps[1],
				// 		ps[2],
				// 		ps[3],
				// 		ps[4],
				// 		t.RateMean(),
				// 		// t.Rate1(),
				// 		// t.Rate5(),
				// 		// t.Rate15(),
				// 	))
			}
			return nil
		}, minLock)
	}
}
