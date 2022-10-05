package metrics

import (
	"sort"
	"strconv"
	"sync"
)

// A HistogramInterface is some strped (no Weights{}, it's not need in registry Each iterator) version of Histogram interface
type HistogramInterface interface {
	Values() []uint64
	Names() []string
	NameTotal() string
	// Tag aliases values (for le key)
	WeightsAliases() []string
}

// A Histogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type Histogram interface {
	HistogramInterface
	Snapshot() Histogram
	Clear()
	Add(v int64)
	Weights() []int64
}

// GetOrRegisterHistoram returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFixedHistogram(name string, r Registry, startVal, endVal, width int64, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedHistogram(startVal, endVal, width, "", "")
	}).(Histogram)
}

// GetOrRegisterHistoramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedHistogram(startVal, endVal, width, "", "")
	}).(Histogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredFixedHistogram(name string, r Registry, startVal, endVal, width int64, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedHistogram(startVal, endVal, width, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedHistogram(startVal, endVal, width, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVHistogram(name string, r Registry, weights []int64, names []string, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVHistogram(weights, names, prefix, total)
	}).(Histogram)
}

func GetOrRegisterVHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVHistogram(weights, names, prefix, total)
	}).(Histogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredVHistogram(name string, r Registry, weights []int64, names []string, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVHistogram(weights, names, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredVHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string, prefix, total string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVHistogram(weights, names, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

type HistogramSnapshot struct {
	weights        []int64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *HistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *HistogramSnapshot) Names() []string {
	return h.names
}

func (h *HistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *HistogramSnapshot) Weights() []int64 {
	return h.weights
}

func (h *HistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *HistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *HistogramSnapshot) Add(v int64) {
	panic("Add called on a HistogramSnapshot")
}

func (h *HistogramSnapshot) Clear() {
	panic("Clear called on a HistogramSnapshot")
}

func (h *HistogramSnapshot) Snapshot() Histogram {
	return h
}

type HistogramStorage struct {
	weights        []int64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *HistogramStorage) Names() []string {
	return h.names
}

func (h *HistogramStorage) NameTotal() string {
	return h.total
}

func (h *HistogramStorage) Weights() []int64 {
	return h.weights
}

func (h *HistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *HistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *HistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *HistogramStorage) Snapshot() Histogram {
	return &HistogramSnapshot{
		names:          h.names,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *HistogramStorage) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

// A FixedHistogram is implementation of Histogram with fixed-size buckets.
type FixedHistogram struct {
	HistogramStorage
	start int64
	width int64
}

func NewFixedHistogram(startVal, endVal, width int64, prefix, total string) *FixedHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	if width < 0 {
		width = -width
	}
	n := endVal - startVal
	if total == "" {
		total = "total"
	}

	count := n/width + 2
	if n%width != 0 {
		count++
	}
	weights := make([]int64, count)
	weightsAliases := make([]string, count)
	names := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal
	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(endVal+width, 10)))
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = ge
			names[i] = prefix + "inf"
			weightsAliases[i] = "inf"
		} else {
			weights[i] = ge
			weightsAliases[i] = strconv.FormatInt(ge, 10)
			names[i] = prefix + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, ge)
			ge += width
		}
	}

	return &FixedHistogram{
		HistogramStorage: HistogramStorage{
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

func (h *FixedHistogram) Add(v int64) {
	var n int64
	if v > h.start {
		n = v - h.start
		if n%h.width == 0 {
			n /= h.width
		} else {
			n = n/h.width + 1
		}
		if n >= int64(len(h.buckets)) {
			n = int64(len(h.buckets) - 1)
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

// A VHistogram is implementation of Histogram with varibale-size buckets.
type VHistogram struct {
	HistogramStorage
}

func NewVHistogram(weights []int64, names []string, prefix, total string) *VHistogram {
	w := make([]int64, len(weights)+1)
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
			if i >= len(names) || names[i] == "" {
				ns[i] = prefix + "inf"
			} else {
				ns[i] = prefix + names[i]
			}
			weightsAliases[i] = "inf"
			w[i] = last
		} else {
			weightsAliases[i] = strconv.FormatInt(w[i], 10)
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				ns[i] = prefix + weightsAliases[i]
			} else {
				ns[i] = prefix + names[i]
			}
		}
	}

	return &VHistogram{
		HistogramStorage: HistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			names:          ns,
			total:          total,
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *VHistogram) Interface() HistogramInterface {
	return h
}

func (h *VHistogram) Snapshot() Histogram {
	return &HistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.names,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VHistogram) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

func (h *VHistogram) Add(v int64) {
	n := searchInt64Ge(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}
