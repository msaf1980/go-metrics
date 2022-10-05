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

type ValTagged struct {
	I       interface{}
	TagsMap map[string]string
}

// A Registry holds references to a set of metrics by name and can iterate
// over them, calling callback functions provided by the user.
//
// This is an interface so as to encourage other structs to implement
// the Registry API as appropriate.
type Registry interface {

	// Call the given function for each registered metric.
	Each(f func(name string, tags string, tagsMap map[string]string, i interface{}) error, minLock bool) error

	// Get the metric by the given name or nil if none is registered.
	Get(name string) interface{}

	// Get the metric by the given name or nil if none is registered.
	GetT(name string, tagsMap map[string]string) interface{}

	// Get an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegister(name string, i interface{}) interface{}

	// Get get an existing metric or registers the given one.
	// The interface can be the metric to register if not found in registry,
	// or a function returning the metric for lazy instantiation.
	GetOrRegisterT(name string, tagsMap map[string]string, i interface{}) interface{}

	// Register the given metric under the given name.
	Register(name string, i interface{}) error

	// Register the given metric under the given name.
	RegisterT(name string, tagsMap map[string]string, i interface{}) error

	// Run all registered healthchecks.
	RunHealthchecks()

	// Unregister the metric with the given name.
	Unregister(name string)

	// Unregister the metric with the given name.
	UnregisterT(name string, tagsMap map[string]string)

	// Unregister all metrics.  (Mostly for testing.)
	UnregisterAll()
}

// The standard implementation of a Registry is a mutex-protected map
// of names to metrics.
type StandardRegistry struct {
	metrics  map[string]interface{}
	metricsT map[NameTagged]*ValTagged
	mutex    sync.RWMutex
}

// Create a new registry.
func NewRegistry() Registry {
	return &StandardRegistry{
		metrics:  make(map[string]interface{}),
		metricsT: make(map[NameTagged]*ValTagged),
	}
}

func (r *StandardRegistry) Each(f func(string, string, map[string]string, interface{}) error, minLock bool) error {
	if minLock {
		return r.each(f)
	}
	return r.eachL(f)
}

// Call the given function for each registered metric.
func (r *StandardRegistry) eachL(f func(string, string, map[string]string, interface{}) error) error {
	var err error
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for name, v := range r.metrics {
		if err = f(name, "", nil, v); err != nil {
			return err
		}
	}
	for k, v := range r.metricsT {
		if err = f(k.Name, k.Tags, v.TagsMap, v.I); err != nil {
			return err
		}
	}
	return nil
}

// Call the given function for each registered metric, minimize locking time with registry copy.
func (r *StandardRegistry) each(f func(string, string, map[string]string, interface{}) error) error {
	var err error
	metrics, metricsT := r.registered()
	for name, v := range metrics {
		if err = f(name, "", nil, v); err != nil {
			return err
		}
	}
	for k, v := range metricsT {
		if err = f(k.Name, k.Tags, v.TagsMap, v.I); err != nil {
			return err
		}
	}
	return nil
}

// Get the metric by the given name or nil if none is registered.
func (r *StandardRegistry) Get(name string) interface{} {
	r.mutex.RLock()
	metric, _ := r.metrics[name]
	r.mutex.RUnlock()
	return metric
}

func (r *StandardRegistry) GetT(name string, tagsMap map[string]string) interface{} {
	tags := JoinTags(tagsMap)
	r.mutex.RLock()
	metric, _ := r.metricsT[NameTagged{Name: name, Tags: tags}]
	r.mutex.RUnlock()
	return metric.I
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
func (r *StandardRegistry) GetOrRegisterT(name string, tagsMap map[string]string, i interface{}) interface{} {
	tags := JoinTags(tagsMap)
	ntags := NameTagged{Name: name, Tags: tags}
	// access the read lock first which should be re-entrant
	r.mutex.RLock()
	metric, ok := r.metricsT[ntags]
	r.mutex.RUnlock()
	if ok {
		return metric.I
	}

	// only take the write lock if we'll be modifying the metrics map
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	metric, ok = r.metricsT[ntags]
	if ok {
		return metric.I
	}
	r.registerT(ntags, &ValTagged{I: i, TagsMap: tagsMap})
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
func (r *StandardRegistry) RegisterT(name string, tagsMap map[string]string, i interface{}) error {
	tags := JoinTags(tagsMap)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// TODO: add tests
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		i = v.Call(nil)[0].Interface()
	}
	return r.registerT(NameTagged{Name: name, Tags: tags}, &ValTagged{I: i, TagsMap: tagsMap})
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
		if h, ok := i.I.(Healthcheck); ok {
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
	r.Each(func(name, tags string, tagsMap map[string]string, i interface{}) error {
		values := make(map[string]interface{})
		switch metric := i.(type) {
		case Counter:
			values["count"] = metric.Count()
		case DownCounter:
			values["count"] = metric.Count()
		case Gauge:
			values["value"] = metric.Value()
		case GaugeFloat64:
			values["value"] = metric.Value()
		case Healthcheck:
			metric.Check()
			values["status"] = metric.IsUp()
		case HistogramInterface:
			var total uint64
			vals := metric.Values()
			for i, name := range metric.Names() {
				values[name] = vals[i]
				total += vals[i]
			}
			values[metric.NameTotal()] = total
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
		return nil
	}, true)
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
func (r *StandardRegistry) UnregisterT(name string, tagsMap map[string]string) {
	tags := JoinTags(tagsMap)
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

func (r *StandardRegistry) registered() (map[string]interface{}, map[NameTagged]*ValTagged) {
	var (
		metrics  map[string]interface{}
		metricsT map[NameTagged]*ValTagged
	)
	r.mutex.RLock()
	if len(r.metrics) > 0 {
		metrics = make(map[string]interface{}, len(r.metrics))
	}
	if len(r.metricsT) > 0 {
		metricsT = make(map[NameTagged]*ValTagged, len(r.metricsT))
	}
	for name, i := range r.metrics {
		metrics[name] = i
	}
	for nameTag, v := range r.metricsT {
		metricsT[nameTag] = v
	}
	r.mutex.RUnlock()
	return metrics, metricsT
}

func (r *StandardRegistry) register(name string, i interface{}) error {
	if _, ok := r.metrics[name]; ok {
		return DuplicateMetric(name)
	}
	switch i.(type) {
	case Counter, DownCounter, Gauge, GaugeFloat64, Healthcheck, HistogramInterface:
		// , Histogram, Meter, Timer:
		r.metrics[name] = i
	default:
		return fmt.Errorf("invalid metric '%s': %#v", name, i)
	}
	return nil
}

func (r *StandardRegistry) registerT(ntags NameTagged, v *ValTagged) error {
	if _, ok := r.metricsT[ntags]; ok {
		return DuplicateMetric(ntags.Name + ntags.Tags)
	}
	switch v.I.(type) {
	case Counter, DownCounter, Gauge, GaugeFloat64, Healthcheck, HistogramInterface:
		// , Histogram, Meter, Timer:
		r.metricsT[ntags] = v
	default:
		return fmt.Errorf("invalid metric '%s': %#v", ntags.Name+ntags.Tags, v.I)
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
	if v, ok := r.metricsT[ntags]; ok {
		if s, ok := v.I.(Stoppable); ok {
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
func Each(f func(name, tags string, tagsMap map[string]string, i interface{}) error, minLock bool) error {
	return DefaultRegistry.Each(f, minLock)
}

// Get the metric by the given name or nil if none is registered.
func Get(name string) interface{} {
	return DefaultRegistry.Get(name)
}

// Get the metric by the given name or nil if none is registered.
func GetT(name string, tagsMap map[string]string) interface{} {
	return DefaultRegistry.GetT(name, tagsMap)
}

// Gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
func GetOrRegister(name string, i interface{}) interface{} {
	return DefaultRegistry.GetOrRegister(name, i)
}

// Gets an existing metric or creates and registers a new one. Threadsafe
// alternative to calling Get and Register on failure.
func GetOrRegisterT(name string, tagsMap map[string]string, i interface{}) interface{} {
	return DefaultRegistry.GetOrRegisterT(name, tagsMap, i)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func Register(name string, i interface{}) error {
	return DefaultRegistry.Register(name, i)
}

// Register the given metric under the given name.  Returns a DuplicateMetric
// if a metric by the given name is already registered.
func RegisterT(name string, tagsMap map[string]string, i interface{}) error {
	return DefaultRegistry.RegisterT(name, tagsMap, i)
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
func MustRegisterT(name string, tagsMap map[string]string, i interface{}) {
	if err := RegisterT(name, tagsMap, i); err != nil {
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
func UnregisterT(name string, tagsMap map[string]string) {
	DefaultRegistry.UnregisterT(name, tagsMap)
}
