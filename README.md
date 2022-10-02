go-metrics
==========

Go port of Coda Hale's Metrics library: <https://github.com/dropwizard/metrics>
Based on `https://github.com/rcrowley/go-metrics`
Main changes:
1. Reduce alocations for better perfomance
2. Read-lock during Registry.Each iterator (for avoid copy registry storage)
3. Don't append ".value" postfix and Gauge and GaugeFloat64
4. modified graphite client

Documentation: <http://godoc.org/github.com/msaf1980/go-metrics>.

Usage
-----

Create and update metrics:

```go
c := metrics.NewCounter()
metrics.Register("foo", c)
c.Inc(47)

g := metrics.NewGauge()
metrics.Register("bar", g)
g.Update(47)

r := NewRegistry()
g := metrics.NewRegisteredFunctionalGauge("cache-evictions", r, func() int64 { return cache.getEvictionsCount() })

s := metrics.NewExpDecaySample(1028, 0.015) // or metrics.NewUniformSample(1028)
h := metrics.NewHistogram(s)
metrics.Register("baz", h)
h.Update(47)

m := metrics.NewMeter()
metrics.Register("quux", m)
m.Mark(47)

t := metrics.NewTimer()
metrics.Register("bang", t)
t.Time(func() {})
t.Update(47)
```

Register() is not threadsafe. For threadsafe metric registration use
GetOrRegister:

```go
t := metrics.GetOrRegisterTimer("account.create.latency", nil)
t.Time(func() {})
t.Update(47)
```

**NOTE:** Be sure to unregister short-lived meters and timers otherwise they will
leak memory:

```go
// Will call Stop() on the Meter to allow for garbage collection
metrics.Unregister("quux")
// Or similarly for a Timer that embeds a Meter
metrics.Unregister("bang")
```

Periodically log every metric in human-readable form to standard error:

```go
go metrics.Log(metrics.DefaultRegistry, 5 * time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
```

Periodically log every metric in slightly-more-parseable form to syslog:

```go
w, _ := syslog.Dial("unixgram", "/dev/log", syslog.LOG_INFO, "metrics")
go metrics.Syslog(metrics.DefaultRegistry, 60e9, w)
```

Periodically emit every metric to Graphite using the Graphite client:

```go

import "github.com/msaf1980/go-metrics/graphite"

go graphite.Graphite(metrics.DefaultRegistry, 10e9, "metrics", "127.0.0.1:2003")
```

Maintain all metrics along with expvars at `/debug/metrics`:

This uses the same mechanism as [the official expvar](http://golang.org/pkg/expvar/)
but exposed under `/debug/metrics`, which shows a json representation of all your usual expvars
as well as all your go-metrics.


```go
import "github.com/msaf1980/go-metrics/exp"

exp.Exp(metrics.DefaultRegistry)
```

Installation
------------

```sh
go get github.com/msaf1980/go-metrics
```

Publishing Metrics
------------------

Clients are available for the following destinations:

* Graphite - https://github.com/msaf1980/go-metrics/graphite
* Log - https://github.com/msaf1980/go-metrics/log
* Syslog - https://github.com/msaf1980/go-metrics/syslog
