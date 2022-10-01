package graphite

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-stringutils"
)

// Config provides a container with configuration parameters for
// the Graphite exporter
type Config struct {
	Host           string        `toml:"host" yaml:"host" json:"host"`                                  // Network address to connect to
	FlushInterval  time.Duration `toml:"interval" yaml:"interval" json:"interval"`                      // Flush interval
	DurationUnit   time.Duration `toml:"duration" yaml:"duration" json:"duration"`                      // Time conversion unit for durations
	Prefix         string        `toml:"prefix" yaml:"prefix" json:"prefix"`                            // Prefix to be prepended to metric names
	ConnectTimeout time.Duration `toml:"connect_timeout" yaml:"connect_timeout" json:"connect_timeout"` // Connect timeout
	Timeout        time.Duration `toml:"timeout" yaml:"timeout" json:"timeout"`                         // Write timeout
	Retry          int           `toml:"retry" yaml:"retry" json:"retry"`                               // Reconnect retry count
	BufSize        int           `toml:"buffer" yaml:"buffer" json:"buffer"`                            // Buffer size

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
	if c.BufSize <= 0 {
		c.BufSize = 1024
	}
	if c.Retry <= 0 {
		c.Retry = 1
	}
	c.percentiles = make([]string, 0, len(c.Percentiles))
	for _, p := range c.Percentiles {
		key := strings.Replace(strconv.FormatFloat(p*100.0, 'f', -1, 64), ".", "", 1)
		c.percentiles = append(c.percentiles, "."+key+"-percentile ")
	}
}

type Graphite struct {
	c    *Config
	conn net.Conn
	buf  stringutils.Builder
	stop chan struct{}
	wg   sync.WaitGroup
}

// Graphite is a blocking exporter function which reports metrics in r
// to a graphite server located at addr, flushing them every d duration
// and prepending metric names with prefix.
func New(d time.Duration, prefix string, host string) *Graphite {
	return WithConfig(&Config{
		Host:          host,
		FlushInterval: d,
		DurationUnit:  time.Nanosecond,
		Prefix:        prefix,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	})
}

// WithConfig is a blocking exporter function just like Graphite,
// but it takes a GraphiteConfig instead.
func WithConfig(c *Config) *Graphite {
	setDefaults(c)
	g := &Graphite{c: c}
	g.buf.Grow(c.BufSize)
	return g
}

// Once performs a single submission to Graphite, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to GraphiteWithConfig for custom error handling.
func Once(c *Config, r metrics.Registry) error {
	setDefaults(c)
	g := &Graphite{c: c}
	g.buf.Grow(c.BufSize)
	err := g.send(r)
	g.Close()
	return err
}

func (g *Graphite) Start(r metrics.Registry) {
	g.wg.Add(1)
	g.stop = make(chan struct{})
	go func() {
		var err error
		defer g.wg.Done()
		t := time.NewTicker(g.c.FlushInterval)
	LOOP:
		for {
			select {
			case <-t.C:
				if err = g.send(r); err != nil {
					log.Println(err)
				}
			case <-g.stop:
				break LOOP
			}
		}
		if err = g.Close(); err != nil {
			log.Println(err)
		}
	}()
}

func (g *Graphite) Stop() {
	g.stop <- struct{}{}
	g.wg.Wait()
}

func (g *Graphite) Close() error {
	err := g.flush()
	if g.conn != nil {
		g.conn.Close()
	}
	return err
}

func (g *Graphite) connect() error {
	var err error
	for i := 0; i < g.c.Retry; i++ {
		g.conn, err = net.DialTimeout("tcp", g.c.Host, g.c.ConnectTimeout)
		if nil == err {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

func (g *Graphite) writeIntMetric(name, postfix string, v, ts int64) (err error) {
	if g.c.Prefix != "" {
		g.buf.WriteString(g.c.Prefix)
		g.buf.WriteRune('.')
	}
	g.buf.WriteString(name)
	g.buf.WriteString(postfix)
	g.buf.WriteString(strconv.FormatInt(v, 10))
	g.buf.WriteRune(' ')
	g.buf.WriteString(strconv.FormatInt(ts, 10))
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) writeFloatMetric(name, postfix string, v float64, ts int64) (err error) {
	if g.c.Prefix != "" {
		g.buf.WriteString(g.c.Prefix)
		g.buf.WriteRune('.')
	}
	g.buf.WriteString(name)
	g.buf.WriteString(postfix)
	g.buf.WriteString(strconv.FormatFloat(v, 'f', 2, 64))
	g.buf.WriteRune(' ')
	g.buf.WriteString(strconv.FormatInt(ts, 10))
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) flush() (err error) {
	if g.buf.Len() > 0 {
		if g.conn == nil {
			if err = g.connect(); err != nil {
				return
			}
		}
		g.conn.SetWriteDeadline(time.Now().Add(g.c.Timeout))
		_, err = g.conn.Write(g.buf.Bytes())
		if err != nil {
			if err = g.connect(); err != nil {
				return
			}
			_, err = g.conn.Write(g.buf.Bytes())
		}
		if err == nil {
			g.buf.Reset()
		}
	}

	return
}

func (g *Graphite) send(r metrics.Registry) error {
	now := time.Now().Unix()
	du := float64(g.c.DurationUnit)
	flushSeconds := float64(g.c.FlushInterval) / float64(time.Second)

	if g.buf.Len() > 10*g.c.BufSize {
		g.buf.Reset() // reset (if previouse write fail and buffer len > 10 * buf_size)
	}

	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			count := metric.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			g.writeIntMetric(name, ".count ", count, now)
			// fmt.Fprintf(w, "%s.%s.count_ps %.2f %d\n", c.Prefix, name, float64(count)/flushSeconds, now)
			g.writeFloatMetric(name, ".count_ps ", float64(count)/flushSeconds, now)
		case metrics.Gauge:
			// fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
			g.writeIntMetric(name, ".value ", metric.Value(), now)
		case metrics.GaugeFloat64:
			// fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
			g.writeFloatMetric(name, ".value ", metric.Value(), now)
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(g.c.Percentiles)
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
			g.writeIntMetric(name, ".count ", h.Count(), now)
			// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
			g.writeIntMetric(name, ".min ", h.Min(), now)
			// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
			g.writeIntMetric(name, ".max ", h.Max(), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
			g.writeFloatMetric(name, ".mean ", h.Mean(), now)
			// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
			g.writeFloatMetric(name, ".std-dev ", h.StdDev(), now)
			for psIdx, psKey := range g.c.percentiles {
				// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
				g.writeFloatMetric(name, psKey, ps[psIdx], now)
			}
		case metrics.Meter:
			m := metric.Snapshot()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
			g.writeIntMetric(name, ".count ", m.Count(), now)
			// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
			g.writeFloatMetric(name, ".one-minute ", m.Rate1(), now)
			// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
			g.writeFloatMetric(name, ".five-minute ", m.Rate5(), now)
			// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
			g.writeFloatMetric(name, ".fifteen-minute ", m.Rate15(), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
			g.writeFloatMetric(name, ".mean ", m.RateMean(), now)
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(g.c.Percentiles)
			count := t.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			g.writeIntMetric(name, ".count ", count, now)
			// fmt.Fprintf(w, "%s.%s.count_ps %.2f %d\n", c.Prefix, name, float64(count)/flushSeconds, now)
			g.writeFloatMetric(name, ".count_ps ", float64(count)/flushSeconds, now)
			// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, t.Min()/int64(du), now)
			g.writeIntMetric(name, ".min ", t.Min()/int64(du), now)
			// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, t.Max()/int64(du), now)
			g.writeIntMetric(name, ".max ", t.Max()/int64(du), now)
			// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, t.Mean()/du, now)
			g.writeFloatMetric(name, ".mean ", t.Mean()/du, now)
			// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, t.StdDev()/du, now)
			g.writeFloatMetric(name, ".std-dev ", t.StdDev()/du, now)
			for psIdx, psKey := range g.c.percentiles {
				// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx]/du, now)
				g.writeFloatMetric(name, psKey, ps[psIdx]/du, now)
			}
			// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, t.Rate1(), now)
			g.writeFloatMetric(name, ".one-minute ", t.Rate1(), now)
			// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, t.Rate5(), now)
			g.writeFloatMetric(name, ".five-minute ", t.Rate5(), now)
			// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, t.Rate15(), now)
			g.writeFloatMetric(name, ".fifteen-minute ", t.Rate15(), now)
			// fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d\n", c.Prefix, name, t.RateMean(), now)
			g.writeFloatMetric(name, ".mean-rate ", t.RateMean(), now)
		default:
			log.Printf("unable to record metric of type %T\n", i)
		}
	})
	return g.flush()
}
