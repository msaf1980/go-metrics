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
func Syslog(r metrics.Registry, d time.Duration, w *syslog.Writer) {
	for range time.Tick(d) {
		r.Each(func(name, tags string, i interface{}) {
			switch metric := i.(type) {
			case metrics.Counter:
				w.Info(fmt.Sprintf("counter %s%s count: %d", name, tags, metric.Count()))
			case metrics.Gauge:
				w.Info(fmt.Sprintf("gauge %s%s value: %d", name, tags, metric.Value()))
			case metrics.GaugeFloat64:
				w.Info(fmt.Sprintf("gauge %s%s: value: %f", name, tags, metric.Value()))
			case metrics.Healthcheck:
				metric.Check()
				w.Info(fmt.Sprintf("healthcheck %s%s error: %v", name, tags, metric.Error()))
			case metrics.Histogram:
				h := metric.Snapshot()
				ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				w.Info(fmt.Sprintf(
					"histogram %s%s count: %d min: %d max: %d mean: %.2f stddev: %.2f median: %.2f 75%%: %.2f 95%%: %.2f 99%%: %.2f 99.9%%: %.2f",
					name, tags,
					h.Count(),
					h.Min(),
					h.Max(),
					h.Mean(),
					h.StdDev(),
					ps[0],
					ps[1],
					ps[2],
					ps[3],
					ps[4],
				))
			case metrics.Meter:
				m := metric.Snapshot()
				w.Info(fmt.Sprintf(
					"meter %s%s count: %d 1-min: %.2f 5-min: %.2f 15-min: %.2f mean: %.2f",
					name, tags,
					m.Count(),
					m.Rate1(),
					m.Rate5(),
					m.Rate15(),
					m.RateMean(),
				))
			case metrics.Timer:
				t := metric.Snapshot()
				ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				w.Info(fmt.Sprintf(
					"timer %s%s count: %d min: %d max: %d mean: %.2f stddev: %.2f median: %.2f 75%%: %.2f 95%%: %.2f 99%%: %.2f 99.9%%: %.2f mean-rate: %.2f\n",
					// " 1-min: %.2f 5-min: %.2f 15-min: %.2f",
					name, tags,
					t.Count(),
					t.Min(),
					t.Max(),
					t.Mean(),
					t.StdDev(),
					ps[0],
					ps[1],
					ps[2],
					ps[3],
					ps[4],
					t.RateMean(),
					// t.Rate1(),
					// t.Rate5(),
					// t.Rate15(),
				))
			}
		})
	}
}
