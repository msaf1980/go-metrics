package graphite

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/msaf1980/go-metrics"
)

// Config provides a container with configuration parameters for
// the Graphite exporter
type Config struct {
	Host           string        `toml:"host" yaml:"host" json:"host"`                                  // Network address to connect to
	FlushInterval  time.Duration `toml:"interval" yaml:"interval" json:"interval"`                      // Flush interval
	DurationUnit   time.Duration `toml:"duration" yaml:"duration" json:"duration"`                      // Time conversion unit for durations
	Prefix         string        `toml:"prefix" yaml:"prefix" json:"prefix"`                            // Prefix to be prepended to metric names
	ConnectTimeout time.Duration `toml:"connect_timeout" yaml:"connect_timeout" json:"connect_timeout"` // Connect timeout
	Timeout        time.Duration `toml:"timeout" yaml:"timeout" json:"timeout"`                         // write timeout

	Percentiles []float64 `toml:"percentiles" yaml:"percentiles" json:"percentiles"` // Percentiles to export from timers and histograms
	percentiles []string  `toml:"-" yaml:"-" json:"-"`                               // Percentiles keys (pregenerated)
}

func setDefaults(c *Config) {
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = time.Minute
	}
	if c.DurationUnit <= 0 {
		c.DurationUnit = time.Millisecond
	}
	c.percentiles = make([]string, 0, len(c.Percentiles))
	for _, p := range c.Percentiles {
		key := strings.Replace(strconv.FormatFloat(p*100.0, 'f', -1, 64), ".", "", 1)
		c.percentiles = append(c.percentiles, "."+key+"-percentile ")
	}
}

// Graphite is a blocking exporter function which reports metrics in r
// to a graphite server located at addr, flushing them every d duration
// and prepending metric names with prefix.
func Graphite(r metrics.Registry, d time.Duration, prefix string, host string) {
	WithConfig(&Config{
		Host:          host,
		FlushInterval: d,
		DurationUnit:  time.Nanosecond,
		Prefix:        prefix,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	}, r)
}

// WithConfig is a blocking exporter function just like Graphite,
// but it takes a GraphiteConfig instead.
func WithConfig(c *Config, r metrics.Registry) {
	setDefaults(c)
	for range time.Tick(c.FlushInterval) {
		if err := graphite(c, r); nil != err {
			log.Println(err)
		}
	}
}

// Once performs a single submission to Graphite, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to GraphiteWithConfig for custom error handling.
func Once(c *Config, r metrics.Registry) error {
	setDefaults(c)
	return graphite(c, r)
}

func writeInt(w *bufio.Writer, prefix, name, postfix string, v, ts int64) (err error) {
	if prefix != "" {
		w.WriteString(prefix)
		w.WriteRune('.')
	}
	w.WriteString(name)
	w.WriteString(postfix)
	w.WriteString(strconv.FormatInt(v, 10))
	w.WriteRune(' ')
	w.WriteString(strconv.FormatInt(ts, 10))
	w.WriteRune('\n')
	return nil
}

func writeFloat(w *bufio.Writer, prefix, name, postfix string, v float64, ts int64) (err error) {
	if prefix != "" {
		w.WriteString(prefix)
		w.WriteRune('.')
	}
	w.WriteString(name)
	w.WriteString(postfix)
	w.WriteString(strconv.FormatFloat(v, 'f', 2, 64))
	w.WriteRune(' ')
	w.WriteString(strconv.FormatInt(ts, 10))
	w.WriteRune('\n')
	return nil
}

func graphiteSend(c *Config, r metrics.Registry, w *bufio.Writer) error {
	now := time.Now().Unix()
	du := float64(c.DurationUnit)
	flushSeconds := float64(c.FlushInterval) / float64(time.Second)

	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			count := metric.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			writeInt(w, c.Prefix, name, ".count ", count, now)
			// fmt.Fprintf(w, "%s.%s.count_ps %.2f %d\n", c.Prefix, name, float64(count)/flushSeconds, now)
			writeFloat(w, c.Prefix, name, ".count_ps ", float64(count)/flushSeconds, now)
		case metrics.Gauge:
			// fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
			writeInt(w, c.Prefix, name, ".value ", metric.Value(), now)
		case metrics.GaugeFloat64:
			// fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
			writeFloat(w, c.Prefix, name, ".value ", metric.Value(), now)
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(c.Percentiles)
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
			writeInt(w, c.Prefix, name, ".count ", h.Count(), now)
			// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
			writeInt(w, c.Prefix, name, ".min ", h.Min(), now)
			// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
			writeInt(w, c.Prefix, name, ".max ", h.Max(), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
			writeFloat(w, c.Prefix, name, ".mean ", h.Mean(), now)
			// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
			writeFloat(w, c.Prefix, name, ".std-dev ", h.StdDev(), now)
			for psIdx, psKey := range c.percentiles {
				// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
				writeFloat(w, c.Prefix, name, psKey, ps[psIdx], now)
			}
		case metrics.Meter:
			m := metric.Snapshot()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
			writeInt(w, c.Prefix, name, ".count ", m.Count(), now)
			// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
			writeFloat(w, c.Prefix, name, ".one-minute ", m.Rate1(), now)
			// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
			writeFloat(w, c.Prefix, name, ".five-minute ", m.Rate5(), now)
			// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
			writeFloat(w, c.Prefix, name, ".fifteen-minute ", m.Rate15(), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
			writeFloat(w, c.Prefix, name, ".mean ", m.RateMean(), now)
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(c.Percentiles)
			count := t.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			writeInt(w, c.Prefix, name, ".count ", count, now)
			// fmt.Fprintf(w, "%s.%s.count_ps %.2f %d\n", c.Prefix, name, float64(count)/flushSeconds, now)
			writeFloat(w, c.Prefix, name, ".count_ps ", float64(count)/flushSeconds, now)
			// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, t.Min()/int64(du), now)
			writeInt(w, c.Prefix, name, ".min ", t.Min()/int64(du), now)
			// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, t.Max()/int64(du), now)
			writeInt(w, c.Prefix, name, ".max ", t.Max()/int64(du), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, t.Mean()/du, now)
			writeFloat(w, c.Prefix, name, ".mean ", t.Mean()/du, now)
			// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, t.StdDev()/du, now)
			writeFloat(w, c.Prefix, name, ".std-dev ", t.StdDev()/du, now)
			for psIdx, psKey := range c.percentiles {
				// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx]/du, now)
				writeFloat(w, c.Prefix, name, psKey, ps[psIdx]/du, now)
			}
			// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, t.Rate1(), now)
			writeFloat(w, c.Prefix, name, ".one-minute ", t.Rate1(), now)
			// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, t.Rate5(), now)
			writeFloat(w, c.Prefix, name, ".five-minute ", t.Rate5(), now)
			// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, t.Rate15(), now)
			writeFloat(w, c.Prefix, name, ".fifteen-minute ", t.Rate15(), now)
			// fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d\n", c.Prefix, name, t.RateMean(), now)
			writeFloat(w, c.Prefix, name, ".mean-rate ", t.RateMean(), now)
		default:
			log.Printf("unable to record metric of type %T\n", i)
		}
		w.Flush()
	})
	return nil
}

func graphite(c *Config, r metrics.Registry) error {
	conn, err := net.DialTimeout("tcp", c.Host, c.ConnectTimeout)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	return graphiteSend(c, r, w)
}
