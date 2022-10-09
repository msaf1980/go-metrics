package metrics

import (
	"time"

	"github.com/msaf1980/go-metrics"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

// Log outputs each metric in the given registry periodically using the given logger.
func Log(r metrics.Registry, freq time.Duration, l Logger, minLock bool) {
	LogScaled(r, freq, time.Nanosecond, l, minLock)
}

// LogOnCue outputs each metric in the given registry on demand through the channel
// using the given logger
func LogOnCue(r metrics.Registry, ch chan interface{}, l Logger, minLock bool) {
	LogScaledOnCue(r, ch, time.Nanosecond, l, minLock)
}

// LogScaled outputs each metric in the given registry periodically using the given
// logger. Print timings in `scale` units (eg time.Millisecond) rather than nanos.
func LogScaled(r metrics.Registry, freq time.Duration, scale time.Duration, l Logger, minLock bool) {
	ch := make(chan interface{})
	go func(channel chan interface{}) {
		for range time.Tick(freq) {
			channel <- struct{}{}
		}
	}(ch)
	LogScaledOnCue(r, ch, scale, l, minLock)
}

// LogScaledOnCue outputs each metric in the given registry on demand through the channel
// using the given logger. Print timings in `scale` units (eg time.Millisecond) rather
// than nanos.
func LogScaledOnCue(r metrics.Registry, ch chan interface{}, scale time.Duration, l Logger, minLock bool) {
	// du := float64(scale)
	// duSuffix := scale.String()[1:]

	for range ch {
		r.Each(func(name, tags string, tagsMap map[string]string, i interface{}) error {
			switch metric := i.(type) {
			case metrics.Counter:
				l.Printf("counter %s%s count: %9d\n", name, tags, metric.Count())
			case metrics.DownCounter:
				l.Printf("counter %s%s count: %9d\n", name, tags, metric.Count())
			case metrics.Gauge:
				l.Printf("gauge %s%s value: %9d\n", name, tags, metric.Value())
			case metrics.UGauge:
				l.Printf("gauge %s%s value: %9d\n", name, tags, metric.Value())
			case metrics.FGauge:
				l.Printf("gauge %s%s value: %f\n", name, tags, metric.Value())
			case metrics.Healthcheck:
				l.Printf("healthcheck %s%s up: %d\n", name, tags, metric.Check())
			case metrics.HistogramInterface:
				vals := metric.Values()
				if metric.IsSummed() {
					for i, label := range metric.Labels() {
						l.Printf("histogram %s%s %s value: %9d\n", name, tags, label, vals[i])
					}
					l.Printf("histogram %s%s %s: %9d\n", name, tags, metric.NameTotal(), vals[0])
				} else {
					var total uint64
					for i, label := range metric.Labels() {
						l.Printf("histogram %s%s %s value: %9d\n", name, tags, label, vals[i])
						total += vals[i]
					}
					l.Printf("histogram %s%s %s: %9d\n", name, tags, metric.NameTotal(), total)
				}
			case metrics.Rate:
				v, rate := metric.Values()
				l.Printf("rate %s%s%s value: %f rate: %f\n", name, metric.Name(), tags, v, rate)
				l.Printf("rate %s%s%s value: %f rate: %f\n", name, metric.RateName(), tags, v, rate)
			case metrics.FRate:
				v, rate := metric.Values()
				l.Printf("rate %s%s%s value: %f rate: %f\n", name, metric.Name(), tags, v, rate)
				l.Printf("rate %s%s%s value: %f rate: %f\n", name, metric.RateName(), tags, v, rate)
				// case metrics.Histogram:
				// 	h := metric.Snapshot()
				// 	ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				// 	l.Printf("histogram %s%s  count: %9d min: %9d max: %9d mean: %9d stddev: %12.2f "+
				// 		"median: %12.2f 75%%: %12.2f 95%%: %12.2f 99%%: %12.2f 99.9%%: %12.2f\n",
				// 		name, tags, h.Count(), h.Min(), h.Max(), h.Mean(), h.StdDev(),
				// 		ps[0], ps[1], ps[2], ps[3], ps[4],
				// 	)
				// case metrics.Meter:
				// 	m := metric.Snapshot()
				// 	l.Printf("meter %s%s  count: %9d 1-min rate: %12.2f 5-min rate: %12.2f 15-min rate: %12.2f mean rate: %12.2f\n",
				// 		name, tags, m.Count(), m.Rate1(), m.Rate5(), m.Rate15(), m.RateMean(),
				// 	)
				// case metrics.Timer:
				// 	t := metric.Snapshot()
				// 	ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
				// 	l.Printf("timer %s%s  count: %9d min: %12.2f%s max: %12.2f%s mean: %12.2f%s stddev: %12.2f%s "+
				// 		"median: %12.2f%s 75%%: %12.2f%s 95%%: %12.2f%s 99%%: %12.2f%s 99.9%%: %12.2f%s "+
				// 		"mean rate: %12.2f\n",
				// 		// " 1-min rate: %12.2f 5-min rate: %12.2 15-min rate: %12.2f\n",
				// 		name, tags, t.Count(), float64(t.Min())/du, duSuffix, float64(t.Max())/du, duSuffix,
				// 		t.Mean()/du, duSuffix, t.StdDev()/du, duSuffix,
				// 		ps[0]/du, duSuffix, ps[1]/du, duSuffix, ps[2]/du, duSuffix, ps[3]/du, duSuffix, ps[4]/du, duSuffix,
				// 		t.RateMean(),
				// 		// t.Rate1(), t.Rate5(), t.Rate15(),
				// 	)
			}
			return nil
		}, minLock)
	}
}
