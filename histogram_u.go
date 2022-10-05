package metrics

import (
	"sort"
	"strconv"
	"sync"
)

// A UHistogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type UHistogram interface {
	HistogramInterface
	Snapshot() UHistogram
	Clear()
	Add(v uint64)
	Weights() []uint64
}

// GetOrRegisterHistoram returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterUFixedHistogram(name string, r Registry, startVal, endVal, width uint64, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewUFixedHistogram(startVal, endVal, width, "", "")
	}).(UHistogram)
}

// GetOrRegisterHistoramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterUFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewUFixedHistogram(startVal, endVal, width, "", "")
	}).(UHistogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredUFixedHistogram(name string, r Registry, startVal, endVal, width uint64, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUFixedHistogram(startVal, endVal, width, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredUFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width uint64, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUFixedHistogram(startVal, endVal, width, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterUVHistogram(name string, r Registry, weights []uint64, names []string, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewUVHistogram(weights, names, prefix, total)
	}).(UHistogram)
}

func GetOrRegisterUVHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewUVHistogram(weights, names, prefix, total)
	}).(UHistogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredUVHistogram(name string, r Registry, weights []uint64, names []string, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUVHistogram(weights, names, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredUVHistogramT(name string, tagsMap map[string]string, r Registry, weights []uint64, names []string, prefix, total string) UHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewUVHistogram(weights, names, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

type UHistogramSnapshot struct {
	weights        []uint64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *UHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *UHistogramSnapshot) Names() []string {
	return h.names
}

func (h *UHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *UHistogramSnapshot) Weights() []uint64 {
	return h.weights
}

func (h *UHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *UHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *UHistogramSnapshot) Add(v uint64) {
	panic("Add called on a UHistogramSnapshot")
}

func (h *UHistogramSnapshot) Clear() {
	panic("Clear called on a UHistogramSnapshot")
}

func (h *UHistogramSnapshot) Snapshot() UHistogram {
	return h
}

type UHistogramStorage struct {
	weights        []uint64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *UHistogramStorage) Names() []string {
	return h.names
}

func (h *UHistogramStorage) NameTotal() string {
	return h.total
}

func (h *UHistogramStorage) Weights() []uint64 {
	return h.weights
}

func (h *UHistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *UHistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *UHistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *UHistogramStorage) Snapshot() UHistogram {
	return &UHistogramSnapshot{
		names:          h.names,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *UHistogramStorage) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

// A UFixedHistogram is implementation of UHistogram with fixed-size buckets.
type UFixedHistogram struct {
	UHistogramStorage
	start uint64
	width uint64
}

func NewUFixedHistogram(startVal, endVal, width uint64, prefix, total string) *UFixedHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	n := endVal - startVal
	if total == "" {
		total = "total"
	}

	count := n/width + 2
	if n%width != 0 {
		count++
	}
	weights := make([]uint64, count)
	weightsAliases := make([]string, count)
	names := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal
	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(endVal+width, 10)))
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = ge
			weightsAliases[i] = "inf"
			names[i] = prefix + weightsAliases[i]
		} else {
			weights[i] = ge
			weightsAliases[i] = strconv.FormatUint(ge, 10)
			names[i] = prefix + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, ge)
			ge += width
		}
	}

	return &UFixedHistogram{
		UHistogramStorage: UHistogramStorage{
			weights:        weights,
			weightsAliases: weightsAliases,
			names:          names,
			total:          total,
			buckets:        buckets,
		},
		start: startVal,
		width: width,
	}
}

func (h *UFixedHistogram) Add(v uint64) {
	var n uint64
	if v > h.start {
		n = v - h.start
		if n%h.width == 0 {
			n /= h.width
		} else {
			n = n/h.width + 1
		}
		if n >= uint64(len(h.buckets)) {
			n = uint64(len(h.buckets)) - 1
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

// A UVHistogram is implementation of UHistogram with varibale-size buckets.
type UVHistogram struct {
	UHistogramStorage
}

func NewUVHistogram(weights []uint64, names []string, prefix, total string) *UVHistogram {
	w := make([]uint64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	last := w[len(w)-2] + 1
	ns := make([]string, len(w))
	if total == "" {
		total = "total"
	}

	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(last, 10)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			weightsAliases[i] = "inf"
			if i >= len(names) || names[i] == "" {
				ns[i] = prefix + weightsAliases[i]
			} else {
				ns[i] = prefix + names[i]
			}
			w[i] = last
		} else {
			weightsAliases[i] = strconv.FormatUint(w[i], 10)
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				ns[i] = prefix + weightsAliases[i]
			} else {
				ns[i] = prefix + names[i]
			}
		}
	}

	return &UVHistogram{
		UHistogramStorage: UHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			names:          ns,
			total:          total,
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *UVHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *UVHistogram) Snapshot() UHistogram {
	return &UHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.names,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *UVHistogram) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

func (h *UVHistogram) Add(v uint64) {
	n := searchUint64Ge(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}
