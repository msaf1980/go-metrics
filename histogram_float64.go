package metrics

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var reZero = regexp.MustCompile(`^(.)\.0+$`)

// A FHistogram is a lossy data structure used to record the distribution of
// non-normally distributed data (like latency) with a high degree of accuracy
// and a bounded degree of precision.
type FHistogram interface {
	HistogramInterface
	Snapshot() FHistogram
	Clear()
	Add(v float64)
	Weights() []float64
}

// GetOrRegisterFHistoram returns an existing FHistogram or constructs and registers
// a new FFixedHistorgam.
func GetOrRegisterFFixedHistogram(name string, r Registry, startVal, endVal, width float64, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFFixedHistogram(startVal, endVal, width, "", "")
	}).(FHistogram)
}

// GetOrRegisterHistoramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam.
func GetOrRegisterFFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFFixedHistogram(startVal, endVal, width, "", "")
	}).(FHistogram)
}

// NewRegisteredFixedHistogram constructs and registers a new FixedHistogram.
func NewRegisteredFFixedHistogram(name string, r Registry, startVal, endVal, width float64, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFFixedHistogram(startVal, endVal, width, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedHistogramT constructs and registers a new FixedHistogram.
func NewRegisteredFFixedHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width float64, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFFixedHistogram(startVal, endVal, width, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterFVHistogram(name string, r Registry, weights []float64, names []string, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFVHistogram(weights, names, prefix, total)
	}).(FHistogram)
}

func GetOrRegisterFVHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFVHistogram(weights, names, prefix, total)
	}).(FHistogram)
}

// NewRegisteredVHistogram constructs and registers a new VHistogram.
func NewRegisteredFVHistogram(name string, r Registry, weights []float64, names []string, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFVHistogram(weights, names, prefix, total)
	r.Register(name, h)
	return h
}

// NewRegisteredVHistogramT constructs and registers a new VHistogram.
func NewRegisteredFVHistogramT(name string, tagsMap map[string]string, r Registry, weights []float64, names []string, prefix, total string) FHistogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFVHistogram(weights, names, prefix, total)
	r.RegisterT(name, tagsMap, h)
	return h
}

type FHistogramSnapshot struct {
	weights        []float64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *FHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *FHistogramSnapshot) Names() []string {
	return h.names
}

func (h *FHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *FHistogramSnapshot) Weights() []float64 {
	return h.weights
}

func (h *FHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *FHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *FHistogramSnapshot) Add(v float64) {
	panic("Add called on a FHistogramSnapshot")
}

func (h *FHistogramSnapshot) Clear() {
	panic("Clear called on a FHistogramSnapshot")
}

func (h *FHistogramSnapshot) Snapshot() FHistogram {
	return h
}

type FHistogramStorage struct {
	weights        []float64 // Sorted weights (greater or equal), last is inf
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last bucket stores endVal overflows count
	lock           sync.RWMutex
}

func (h *FHistogramStorage) Names() []string {
	return h.names
}

func (h *FHistogramStorage) NameTotal() string {
	return h.total
}

func (h *FHistogramStorage) Weights() []float64 {
	return h.weights
}

func (h *FHistogramStorage) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *FHistogramStorage) Interface() HistogramInterface {
	return h
}

func (h *FHistogramStorage) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *FHistogramStorage) Snapshot() FHistogram {
	return &FHistogramSnapshot{
		names:          h.names,
		total:          h.total,
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		buckets:        h.buckets,
	}
}

func (h *FHistogramStorage) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

// A FFixedHistogram is implementation of FHistogram with fixed-size buckets.
type FFixedHistogram struct {
	FHistogramStorage
	start float64
	width float64
}

func NewFFixedHistogram(startVal, endVal, width float64, prefix, total string) FHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	n := (endVal - startVal) / width
	if n > float64(int(n)) {
		n++
	}
	if total == "" {
		total = "total"
	}

	count := int(n) + 2
	weights := make([]float64, count)
	weightsAliases := make([]string, count)
	names := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal

	// maxLength := len(strconv.FormatInt(int64(endVal+width)+1, 10))
	// fmtStr := fmt.Sprintf("%%s%%0%dd", maxLength)
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = ge
			weightsAliases[i] = "inf"
			names[i] = prefix + weightsAliases[i]
		} else {
			weights[i] = ge
			// n := int(ge)
			// d := ge - float64(n)
			weightsAliases[i] = strings.ReplaceAll(reZero.ReplaceAllString(strconv.FormatFloat(ge, 'f', -1, 64), "${1}"), ".", "_")
			names[i] = prefix + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, n)
			ge += width
		}
	}

	return &FFixedHistogram{
		FHistogramStorage: FHistogramStorage{
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

func (h *FFixedHistogram) Add(v float64) {
	var (
		n int
		f float64
	)
	if v > h.start {
		f = (v - h.start) / h.width
		if f > float64(int(f)) {
			n = int(f) + 1
		} else {
			n = int(f)
		}
		if n >= len(h.buckets) {
			n = len(h.buckets) - 1
		}
	}
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}

// A FVHistogram is implementation of FHistogram with varibale-size buckets.
type FVHistogram struct {
	FHistogramStorage
}

func NewFVHistogram(weights []float64, names []string, prefix, total string) *FVHistogram {
	w := make([]float64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	last := w[len(w)-2] + 1
	ns := make([]string, len(w))
	if total == "" {
		total = "total"
	}

	// fmtStr := fmt.Sprintf("%%s%%0%df", len(strconv.FormatFloat(last, 'f', -1, 64)))
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
			weightsAliases[i] = strings.ReplaceAll(reZero.ReplaceAllString(strconv.FormatFloat(w[i], 'f', -1, 64), "${1}"), ".", "_")
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				ns[i] = prefix + weightsAliases[i]
			} else {
				ns[i] = prefix + names[i]
			}
		}
	}

	return &FVHistogram{
		FHistogramStorage: FHistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			names:          ns,
			total:          total,
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *FVHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *FVHistogram) Snapshot() FHistogram {
	return &FHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.names,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *FVHistogram) Clear() {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	h.buckets = buckets
	h.lock.Unlock()
}

func (h *FVHistogram) Add(v float64) {
	n := searchFloat64Ge(h.weights, v)
	h.lock.Lock()
	h.buckets[n]++
	h.lock.Unlock()
}
