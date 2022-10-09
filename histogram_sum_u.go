package metrics

import (
	"sort"
	"strconv"
)

// GetOrRegisterHistogram returns an existing Histogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumUHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewFixedSumUHistogram(startVal, endVal, width)
	}).(Histogram)
}

// GetOrRegisterSumUHistogramT returns an existing Histogram or constructs and registers
// a new FixedHistorgam (prometheus-like histogram).
func GetOrRegisterFixedSumUHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewFixedSumUHistogram(startVal, endVal, width)
	}).(Histogram)
}

// NewRegisteredFixedSumUHistogram constructs and registers a new FixedSumUHistogram (prometheus-like histogram).
func NewRegisteredFixedSumUHistogram(name string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumUHistogram(startVal, endVal, width)
	r.Register(name, h)
	return h
}

// NewRegisteredFixedSumUHistogramT constructs and registers a new FixedSumUHistogram (prometheus-like histogram).
func NewRegisteredFixedSumUHistogramT(name string, tagsMap map[string]string, r Registry, startVal, endVal, width int64) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewFixedSumUHistogram(startVal, endVal, width)
	r.RegisterT(name, tagsMap, h)
	return h
}

func GetOrRegisterVSumUHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegister(name, func() interface{} {
		return NewVSumUHistogram(weights, names)
	}).(Histogram)
}

func GetOrRegisterVSumUHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	return r.GetOrRegisterT(name, tagsMap, func() interface{} {
		return NewVSumUHistogram(weights, names)
	}).(Histogram)
}

// NewRegisteredVSumUHistogram constructs and registers a new VSumUHistogram (prometheus-like histogram).
func NewRegisteredVSumUHistogram(name string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumUHistogram(weights, names)
	r.Register(name, h)
	return h
}

// NewRegisteredVSumUHistogramT constructs and registers a new VSumUHistogram (prometheus-like histogram).
func NewRegisteredVSumUHistogramT(name string, tagsMap map[string]string, r Registry, weights []int64, names []string) Histogram {
	if nil == r {
		r = DefaultRegistry
	}
	h := NewVSumUHistogram(weights, names)
	r.RegisterT(name, tagsMap, h)
	return h
}

type SumUHistogramSnapshot struct {
	weights        []int64 // Sorted weights, by <=
	weightsAliases []string
	names          []string
	total          string
	buckets        []uint64 // last buckets stores all, not included at previous
}

func (h *SumUHistogramSnapshot) Values() []uint64 {
	return h.buckets
}

func (h *SumUHistogramSnapshot) Labels() []string {
	return h.names
}

func (SumUHistogramSnapshot) SetLabels([]string) Histogram {
	panic("SetLabels called on a HistogramSnapshot")
}

func (SumUHistogramSnapshot) AddLabelPrefix(string) Histogram {
	panic("AddLabelPrefix called on a SumUHistogramSnapshot")
}
func (SumUHistogramSnapshot) SetNameTotal(string) Histogram {
	panic("SetNameTotal called on a SumUHistogramSnapshot")
}

func (h *SumUHistogramSnapshot) NameTotal() string {
	return h.total
}

func (h *SumUHistogramSnapshot) Weights() []int64 {
	return h.weights
}

func (h *SumUHistogramSnapshot) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *SumUHistogramSnapshot) Interface() HistogramInterface {
	return h
}

func (h *SumUHistogramSnapshot) Add(v int64) {
	panic("Add called on a SumUHistogramSnapshot")
}

func (h *SumUHistogramSnapshot) Clear() []uint64 {
	panic("Clear called on a HistogramSnapshot")
}

func (h *SumUHistogramSnapshot) Snapshot() Histogram {
	return h
}

func (SumUHistogramSnapshot) IsSummed() bool { return true }

// A FixedSumUHistogram is implementation of prometheus-like Histogram with fixed-size buckets.
type FixedSumUHistogram struct {
	HistogramStorage
	start int64
	width int64
}

func NewFixedSumUHistogram(startVal, endVal, width int64) *FixedSumUHistogram {
	if endVal < startVal {
		startVal, endVal = endVal, startVal
	}
	if width < 0 {
		width = -width
	}
	n := endVal - startVal

	count := n/width + 2
	if n%width != 0 {
		count++
	}
	weights := make([]int64, count)
	weightsAliases := make([]string, count)
	labels := make([]string, count)
	buckets := make([]uint64, count)
	ge := startVal
	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(endVal+width, 10)))
	for i := 0; i < len(weights); i++ {
		if i == len(weights)-1 {
			weights[i] = ge
			weightsAliases[i] = "inf"
			labels[i] = ".inf"
		} else {
			weights[i] = ge
			weightsAliases[i] = strconv.FormatInt(ge, 10)
			labels[i] = "." + weightsAliases[i]
			// names[i] = fmt.Sprintf(fmtStr, prefix, ge)
			ge += width
		}
	}

	return &FixedSumUHistogram{
		HistogramStorage: HistogramStorage{
			weights:        weights,
			weightsAliases: weightsAliases,
			labels:         labels,
			total:          ".total",
			buckets:        buckets,
		},
		start: startVal,
		width: width,
	}
}

func (h *FixedSumUHistogram) Add(v int64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *FixedSumUHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *FixedSumUHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *FixedSumUHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}

func (h *FixedSumUHistogram) Clear() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	v := h.buckets
	h.buckets = buckets
	h.lock.Unlock()
	return v
}

func (h *FixedSumUHistogram) IsSummed() bool { return true }

// A VSumUHistogram is implementation of prometheus-like Histogram with varibale-size buckets.
type VSumUHistogram struct {
	HistogramStorage
}

func NewVSumUHistogram(weights []int64, names []string) *VSumUHistogram {
	w := make([]int64, len(weights)+1)
	weightsAliases := make([]string, len(w))
	copy(w, weights)
	sort.Slice(w[:len(weights)-1], func(i, j int) bool { return w[i] < w[j] })
	last := w[len(w)-2] + 1
	lbls := make([]string, len(w))

	// fmtStr := fmt.Sprintf("%%s%%0%dd", len(strconv.FormatUint(last, 10)))
	for i := 0; i < len(w); i++ {
		if i == len(w)-1 {
			if i >= len(names) || names[i] == "" {
				lbls[i] = ".inf"
			} else {
				lbls[i] = names[i]
			}
			weightsAliases[i] = "inf"
			w[i] = last
		} else {
			weightsAliases[i] = strconv.FormatInt(w[i], 10)
			if i >= len(names) || names[i] == "" {
				// ns[i] = fmt.Sprintf(fmtStr, prefix, w[i])
				lbls[i] = "." + weightsAliases[i]
			} else {
				lbls[i] = names[i]
			}
		}
	}

	return &VSumUHistogram{
		HistogramStorage: HistogramStorage{
			weights:        w,
			weightsAliases: weightsAliases,
			labels:         lbls,
			total:          ".total",
			buckets:        make([]uint64, len(w)),
		},
	}
}

func (h *VSumUHistogram) Values() []uint64 {
	buckets := make([]uint64, len(h.buckets))
	h.lock.Lock()
	copy(buckets, h.buckets)
	h.lock.Unlock()
	return buckets
}

func (h *VSumUHistogram) WeightsAliases() []string {
	return h.weightsAliases
}

// for static check compatbility with HistogramInterface
func (h *VSumUHistogram) Interface() HistogramInterface {
	return h
}

func (h *VSumUHistogram) Snapshot() Histogram {
	return &SumUHistogramSnapshot{
		weights:        h.weights,
		weightsAliases: h.weightsAliases,
		names:          h.labels,
		total:          h.NameTotal(),
		buckets:        h.Values(),
	}
}

func (h *VSumUHistogram) Add(v int64) {
	h.lock.Lock()
	for i := 0; i < len(h.buckets); i++ {
		h.buckets[i]++
		if v <= h.weights[i] {
			break
		}
	}
	h.lock.Unlock()
}

func (h *VSumUHistogram) SetLabels(labels []string) Histogram {
	h.HistogramStorage.SetLabels(labels)
	return h
}

func (h *VSumUHistogram) AddLabelPrefix(labelPrefix string) Histogram {
	h.HistogramStorage.AddLabelPrefix(labelPrefix)
	return h
}
func (h *VSumUHistogram) SetNameTotal(total string) Histogram {
	h.HistogramStorage.SetNameTotal(total)
	return h
}

func (h *VSumUHistogram) IsSummed() bool { return true }
