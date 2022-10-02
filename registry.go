package metrics

import (
	"fmt"
	"reflect"
	"sync"
)

// DuplicateMetric is the error returned by Registry.Register when a metric
// already exists.  If you mean to Register that metric you must first
// Unregister the existing metric.
type DuplicateMetric string

func (err DuplicateMetric) Error() string {
	return "duplicate metric: " + string(err)
}

type NameTagged struct {
	Name string
	Tags string
}

// A Registry holds references to a set of metrics by name and can iterate
// over them, calling callback functions provided by the user.
//
// This is an interface so as to encourage other structs to implement
// the Registry API as appropriate.
type Registry interface {

	// Call the given function for each registered metric.
	Each(func(name string, tags string, i interface{}))

	// Get the metric by the given name or nil if none is registered.
	Get(name string) interface{}

	// Get the metric by the given name or nil if none is registered.
	GetT(name string, tags string) interface{}

	// Get an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegister(name string, i interface{}) interface{}

	// Get get an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegisterT(name string, tags string, i interface{}) interface{}

	// Register the given metric under the given name.
	Register(name string, i interface{}) error

	// Register the given metric under the given name.
	RegisterT(name string, tags string, i interface{}) error

	// Run all registered healthchecks.
	RunHealthchecks()

	// Unregister the metric with the given name.
	Unregister(name string)

	// Unregister the metric with the given name.
	UnregisterT(name, tags string)

	// Unregister all metrics.  (Mostly for testing.)
	UnregisterAll()
}

// The standard implementation of a Registry is a mutex-protected map
// of names to metrics.
type StandardRegistry struct {
	metrics  map[string]interface{}
	metricsT map[NameTagged]interface{}
	mutex    sync.RWMutex
}

// Create a new registry.
func NewRegistry() Registry {
	return &StandardRegistry{
		metrics:  make(map[string]interface{}),
		metricsT: make(map[NameTagged]interface{}),
	}
}

// Call the given function for each registered metric.
func (r *StandardRegistry) Each(f func(string, string, interface{})) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for name, v := range r.metrics {
		f(name, "", v)
	}
	for k, v := range r.metricsT {
		f(k.Name, k.Tags, v)
	}
}

// Get the metric by the given name or nil if none is registered.
func (r *StandardRegistry) Get(name string) interface{} {
	r.mutex.RLock()
	metric, _ := r.metrics[name]
	r.mutex.RUnlock()
	return metric
}

// GetT the metric by the given name and tags or nil if none is registered.
func (r *StandardRegistry) GetT(name, tags string) interface{} {
	r.mutex.RLock()
	metric, _ := r.metricsT[NameTagged{Name: name, Tags: tags}]
	r.mutex.RUnlock()
	return metric
}

// Get an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
// The interface can be the metric to register if not found in registry,
// or a function returning the metric for lazy instantiation.
func (r *StandardRegistry) GetOrRegister(name string, i interface{}) interface{} {
	// access the read lock first which should be re-entrant
	r.mutex.RLock()
	metric, ok := r.metrics[name]
	r.mutex.RUnlock()
	if ok {
		return metric
	}

	// only take the write lock if we'll be modifying the metrics map
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	metric, ok = r.metrics[name]
	if ok {
		return metric
	}

	r.register(name, i)
	return i
}

// Get an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
// The interface can be the metric to register if not found in registry,
// or a function returning the metric for lazy instantiation.
func (r *StandardRegistry) GetOrRegisterT(name, tags string, i interface{}) interface{} {
	ntags := NameTagged{Name: name, Tags: tags}
	// access the read lock first which should be re-entrant
	r.mutex.RLock()
	metric, ok := r.metricsT[ntags]
	r.mutex.RUnlock()
	if ok {
		return metric
	}

	// only take the write lock if we'll be modifying the metrics map
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	metric, ok = r.metricsT[ntags]
	if ok {
		return metric
	}

	r.registerT(ntags, i)
	return i
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *StandardRegistry) Register(name string, i interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// TODO: add tests
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}
	return r.register(name, i)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func (r *StandardRegistry) RegisterT(name string, tags string, i interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// TODO: add tests
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}
	return r.registerT(NameTagged{Name: name, Tags: tags}, i)
}

// Run all registered healthchecks.
func (r *StandardRegistry) RunHealthchecks() {
	r.mutex.RLock()
	hs := make([]Healthcheck, 0, len(r.metrics)+len(r.metricsT))
	for _, i := range r.metrics {
		if h, ok := i.(Healthcheck); ok {
			hs = append(hs, h)
		}
	}
	for _, i := range r.metricsT {
		if h, ok := i.(Healthcheck); ok {
			hs = append(hs, h)
		}
	}
	r.mutex.RUnlock()

	for _, h := range hs {
		h.Check()
	}
}

// GetAll metrics in the Registry
func (r *StandardRegistry) GetAll() map[string]map[string]interface{} {
	data := make(map[string]map[string]interface{})
	r.Each(func(name, tags string, i interface{}) {
		values := make(map[string]interface{})
		switch metric := i.(type) {
		case Counter:
			values["count"] = metric.Count()
		case Gauge:
			values["value"] = metric.Value()
		case GaugeFloat64:
			values["value"] = metric.Value()
		case Healthcheck:
			values["error"] = nil
			metric.Check()
			if err := metric.Error(); nil != err {
				values["error"] = metric.Error().Error()
			}
			// case Histogram:
			// 	h := metric.Snapshot()
			// 	ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			// 	values["count"] = h.Count()
			// 	values["min"] = h.Min()
			// 	values["max"] = h.Max()
			// 	values["mean"] = h.Mean()
			// 	values["stddev"] = h.StdDev()
			// 	values["median"] = ps[0]
			// 	values["75%"] = ps[1]
			// 	values["95%"] = ps[2]
			// 	values["99%"] = ps[3]
			// 	values["99.9%"] = ps[4]
			// case Meter:
			// 	m := metric.Snapshot()
			// 	values["count"] = m.Count()
			// 	values["1m.rate"] = m.Rate1()
			// 	values["5m.rate"] = m.Rate5()
			// 	values["15m.rate"] = m.Rate15()
			// 	values["mean.rate"] = m.RateMean()
			// case Timer:
			// 	t := metric.Snapshot()
			// 	ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			// 	values["count"] = t.Count()
			// 	values["min"] = t.Min()
			// 	values["max"] = t.Max()
			// 	values["mean"] = t.Mean()
			// 	values["stddev"] = t.StdDev()
			// 	values["median"] = ps[0]
			// 	values["75%"] = ps[1]
			// 	values["95%"] = ps[2]
			// 	values["99%"] = ps[3]
			// 	values["99.9%"] = ps[4]
			// 	values["1m.rate"] = t.Rate1()
			// 	values["5m.rate"] = t.Rate5()
			// 	values["15m.rate"] = t.Rate15()
			// 	values["mean.rate"] = t.RateMean()
		}
		data[name+tags] = values
	})
	return data
}

// Unregister the metric with the given name.
func (r *StandardRegistry) Unregister(name string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.stop(name)
	delete(r.metrics, name)
}

// Unregister the metric with the given name.
func (r *StandardRegistry) UnregisterT(name, tags string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ntags := NameTagged{Name: name, Tags: tags}
	r.stopT(ntags)
	delete(r.metricsT, ntags)
}

// Unregister all metrics.  (Mostly for testing.)
func (r *StandardRegistry) UnregisterAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name := range r.metrics {
		r.stop(name)
		delete(r.metrics, name)
	}
	for ntags := range r.metricsT {
		r.stopT(ntags)
		delete(r.metricsT, ntags)
	}
}

func (r *StandardRegistry) register(name string, i interface{}) error {
	if _, ok := r.metrics[name]; ok {
		return DuplicateMetric(name)
	}
	switch i.(type) {
	case Counter, Gauge, GaugeFloat64, Healthcheck:
		// , Histogram, Meter, Timer:
		r.metrics[name] = i
	default:
		return fmt.Errorf("invalid metric '%s': %#v", name, i)
	}
	return nil
}

func (r *StandardRegistry) registerT(ntags NameTagged, i interface{}) error {
	if _, ok := r.metricsT[ntags]; ok {
		return DuplicateMetric(ntags.Name + ntags.Tags)
	}
	switch i.(type) {
	case Counter, Gauge, GaugeFloat64, Healthcheck:
		// , Histogram, Meter, Timer:
		r.metricsT[ntags] = i
	default:
		return fmt.Errorf("invalid metric '%s': %#v", ntags.Name+ntags.Tags, i)
	}
	return nil
}

func (r *StandardRegistry) stop(name string) {
	if i, ok := r.metrics[name]; ok {
		if s, ok := i.(Stoppable); ok {
			s.Stop()
		}
	}
}

func (r *StandardRegistry) stopT(ntags NameTagged) {
	if i, ok := r.metricsT[ntags]; ok {
		if s, ok := i.(Stoppable); ok {
			s.Stop()
		}
	}
}

// Stoppable defines the metrics which has to be stopped.
type Stoppable interface {
	Stop()
}

var DefaultRegistry Registry = NewRegistry()

// Call the given function for each registered metric.
func Each(f func(name, tags string, i interface{})) {
	DefaultRegistry.Each(f)
}

// Get the metric by the given name or nil if none is registered.
func Get(name string) interface{} {
	return DefaultRegistry.Get(name)
}

// Get the metric by the given name or nil if none is registered.
func GetT(name, tags string) interface{} {
	return DefaultRegistry.GetT(name, tags)
}

// Gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
func GetOrRegister(name string, i interface{}) interface{} {
	return DefaultRegistry.GetOrRegister(name, i)
}

// Gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
func GetOrRegisterT(name, tags string, i interface{}) interface{} {
	return DefaultRegistry.GetOrRegisterT(name, tags, i)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func Register(name string, i interface{}) error {
	return DefaultRegistry.Register(name, i)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func RegisterT(name, tags string, i interface{}) error {
	return DefaultRegistry.RegisterT(name, tags, i)
}

// Register the given metric under the given name.  Panics if a metric by the
// given name is already registered.
func MustRegister(name string, i interface{}) {
	if err := Register(name, i); err != nil {
		panic(err)
	}
}

// Register the given metric under the given name.  Panics if a metric by the
// given name is already registered.
func MustRegisterT(name, tags string, i interface{}) {
	if err := RegisterT(name, tags, i); err != nil {
		panic(err)
	}
}

// Run all registered healthchecks.
func RunHealthchecks() {
	DefaultRegistry.RunHealthchecks()
}

// Unregister the metric with the given name.
func Unregister(name string) {
	DefaultRegistry.Unregister(name)
}

// Unregister the metric with the given name.
func UnregisterT(name, tags string) {
	DefaultRegistry.UnregisterT(name, tags)
}
