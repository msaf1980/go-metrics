package metrics

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/msaf1980/go-metrics/test"
)

func TestNewFFixedloatHistogram(t *testing.T) {
	tests := []struct {
		startVal           float64
		endVal             float64
		width              float64
		labelPrefix        string
		total              string
		wantWeights        []float64
		wantWeightsAliases []string
		wantLabels         []string
	}{
		{
			startVal:           100,
			endVal:             1000,
			width:              100,
			wantWeights:        []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, math.MaxFloat64},
			wantWeightsAliases: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000", "inf"},
			// wantLabels:   []string{"0100", "0200", "0300", "0400", "0500", "0600", "0700", "0800", "0900", "1000", "inf"},
			wantLabels: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000", "inf"},
		},
		{
			startVal:           10,
			endVal:             100,
			width:              40,
			labelPrefix:        "req_le_",
			wantWeights:        []float64{10, 50, 90, 130, math.MaxFloat64},
			wantWeightsAliases: []string{"10", "50", "90", "130", "inf"},
			wantLabels:         []string{"req_le_10", "req_le_50", "req_le_90", "req_le_130", "req_le_inf"},
		},
		{
			startVal:           10.1,
			endVal:             100,
			width:              40.01,
			labelPrefix:        "req_le_",
			wantWeights:        []float64{10.1, 50.11, 90.12, 130.13, math.MaxFloat64},
			wantWeightsAliases: []string{"10_10", "50_11", "90_12", "130_13", "inf"},
			wantLabels:         []string{"req_le_10_10", "req_le_50_11", "req_le_90_12", "req_le_130_13", "req_le_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewFFixedHistogram(tt.startVal, tt.endVal, tt.width)
			if tt.labelPrefix != "" {
				got.AddLabelPrefix(tt.labelPrefix)
			}
			if tt.total != "" {
				got.SetNameTotal(tt.total)
			}
			if !test.SliceFloatEq(got.Weights(), tt.wantWeights) {
				t.Errorf("NewUFFistogram() weights = %+v, want %+v", got.Weights(), tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.Labels(), tt.wantLabels) {
				t.Errorf("NewUFFistogram() names =\n%q\nwant\n%q", got.Labels(), tt.wantLabels)
			}
			if tt.total == "" {
				tt.total = "total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewUFFistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.Labels()) != len(got.Values()) {
				t.Errorf("NewUFFistogram() buckets count =%d, want %d", len(got.Labels()), len(got.Values()))
			}
		})
	}
}

func TestFFixedHistogram_Add(t *testing.T) {
	startVal := 10.0
	width := 10.0
	endVal := 50.
	h := NewFFixedHistogram(startVal, endVal, width)
	r := NewRegistry()
	if err := r.Register("histogram", h); err != nil {
		t.Error(err)
	}
	tests := []struct {
		add  float64
		want []uint64
	}{
		{add: 0, want: []uint64{1, 0, 0, 0, 0, 0}},
		{add: 10, want: []uint64{2, 0, 0, 0, 0, 0}},
		{add: 11, want: []uint64{2, 1, 0, 0, 0, 0}},
		{add: 20, want: []uint64{2, 2, 0, 0, 0, 0}},
		{add: 25, want: []uint64{2, 2, 1, 0, 0, 0}},
		{add: 49, want: []uint64{2, 2, 1, 0, 1, 0}},
		{add: 50, want: []uint64{2, 2, 1, 0, 2, 0}},
		{add: 51, want: []uint64{2, 2, 1, 0, 2, 1}},
		{add: 100, want: []uint64{2, 2, 1, 0, 2, 2}},
	}

	// zero
	got := h.Values()
	if !reflect.DeepEqual([]uint64{0, 0, 0, 0, 0, 0}, got) {
		t.Errorf("h.Values() = %v, want zero %v", got, []uint64{0, 0, 0, 0, 0, 0})
	}

	for n, tt := range tests {
		t.Run(strconv.FormatFloat(tt.add, 'f', -1, 64)+"#"+strconv.Itoa(n), func(t *testing.T) {
			h.Add(tt.add)
			got := h.Values()
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("h.Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFFixedHistogram_SetNames(t *testing.T) {
	startVal := float64(10.1)
	width := float64(10.2)
	endVal := float64(50)
	h := NewFFixedHistogram(startVal, endVal, width)

	wantLabels := []string{"10_10", "20_30", "30_50", "40_70", "50_90", "inf"}
	weightsAliases := wantLabels
	wantNameTotal := "total"
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"le_10_10", "le_20_30", "le_30_50", "le_40_70", "le_50_90", "le_inf"}
	wantNameTotal = "req_total"
	h.AddLabelPrefix("le_")
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"green", "yellow", "le_30_50", "le_40_70", "le_50_90", "le_inf"}
	h.SetLabels([]string{"green", "yellow"})
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}
}

func TestFFixedHistogram_Snapshot(t *testing.T) {
	startVal := 10.0
	width := 10.0
	endVal := 50.0
	r := NewRegistry()
	h := GetOrRegisterFFixedHistogram("histogram", r, startVal, endVal, width)
	h.AddLabelPrefix("le_")
	h.SetNameTotal("req_total")
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{0, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Labels(), got.Labels()) {
		t.Errorf("h.Snapshot().Labels() = %v, want %v", got.Labels(), h.Labels())
	}
	if got.NameTotal() != h.NameTotal() {
		t.Errorf("NewFFixedHistogram() total = %q, want %q", got.NameTotal(), h.NameTotal())
	}
	if !reflect.DeepEqual(h.Weights(), got.Weights()) {
		t.Errorf("h.Snapshot().Weights() = %v, want %v", got.Weights(), h.Weights())
	}
	if !reflect.DeepEqual(got.WeightsAliases(), h.WeightsAliases()) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), h.WeightsAliases())
	}
}

func TestNewFVHistogram(t *testing.T) {
	tests := []struct {
		weights            []float64
		names              []string
		labelPrefix        string
		total              string
		wantWeights        []float64
		wantWeightsAliases []string
		wantLabels         []string
	}{
		{
			weights:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20},
			wantWeights:        []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20, math.MaxFloat64},
			wantWeightsAliases: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "20", "inf"},
			// wantLabels:   []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "20", "inf"},
			wantLabels: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "20", "inf"},
		},
		{
			weights:            []float64{10, 20, 100},
			names:              []string{"green", "blue", "yellow", "red", "none"},
			wantWeights:        []float64{10, 20, 100, math.MaxFloat64},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantLabels:         []string{"green", "blue", "yellow", "red"},
		},
		{
			weights:            []float64{10, 20, 100},
			names:              []string{"green", "blue", "yellow"},
			labelPrefix:        "req_",
			total:              "total_req",
			wantWeights:        []float64{10, 20, 100, math.MaxFloat64},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantLabels:         []string{"req_green", "req_blue", "req_yellow", "req_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewFVHistogram(tt.weights, tt.names)
			if tt.labelPrefix != "" {
				got.AddLabelPrefix(tt.labelPrefix)
			}
			if tt.total != "" {
				got.SetNameTotal(tt.total)
			}
			if !reflect.DeepEqual(got.weights, tt.wantWeights) {
				t.Errorf("NewFVHistogram() weights = %+v, want %+v", got.weights, tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.labels, tt.wantLabels) {
				t.Errorf("NewFVHistogram() names =\n%q\nwant\n%q", got.labels, tt.wantLabels)
			}
			if tt.total == "" {
				tt.total = "total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewFVHistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.labels) != len(got.Values()) {
				t.Errorf("NewFVHistogram() buckets count =%d, want %d", len(got.Values()), len(tt.wantLabels))
			}
		})
	}
}

func TestFVHistogram_Add(t *testing.T) {
	r := NewRegistry()
	h := GetOrRegisterFVHistogram("histogram", r, []float64{1, 2, 5, 8, 20}, nil)
	tests := []struct {
		add  float64
		want []uint64
	}{
		{add: 0, want: []uint64{1, 0, 0, 0, 0, 0}},
		{add: 1, want: []uint64{2, 0, 0, 0, 0, 0}},
		{add: 2, want: []uint64{2, 1, 0, 0, 0, 0}},
		{add: 3, want: []uint64{2, 1, 1, 0, 0, 0}},
		{add: 5, want: []uint64{2, 1, 2, 0, 0, 0}},
		{add: 6, want: []uint64{2, 1, 2, 1, 0, 0}},
		{add: 7, want: []uint64{2, 1, 2, 2, 0, 0}},
		{add: 8, want: []uint64{2, 1, 2, 3, 0, 0}},
		{add: 9, want: []uint64{2, 1, 2, 3, 1, 0}},
		{add: 20, want: []uint64{2, 1, 2, 3, 2, 0}},
		{add: 21, want: []uint64{2, 1, 2, 3, 2, 1}},
		{add: 100, want: []uint64{2, 1, 2, 3, 2, 2}},
	}

	// zero
	got := h.Values()
	if !reflect.DeepEqual([]uint64{0, 0, 0, 0, 0, 0}, got) {
		t.Errorf("h.Values() = %v, want zero %v", got, []uint64{0, 0, 0, 0, 0, 0})
	}

	for n, tt := range tests {
		t.Run(strconv.FormatFloat(tt.add, 'f', -1, 64)+"#"+strconv.Itoa(n), func(t *testing.T) {
			h.Add(tt.add)
			got := h.Values()
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("h.Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFVHistogram_SetNames(t *testing.T) {
	h := NewFVHistogram([]float64{10, 20.1, 50.34, 80.785, 100}, nil)

	wantLabels := []string{"10", "20_10", "50_34", "80_78", "100", "inf"}
	weightsAliases := wantLabels
	wantNameTotal := "total"
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"le_10", "le_20_10", "le_50_34", "le_80_78", "le_100", "le_inf"}
	wantNameTotal = "req_total"
	h.AddLabelPrefix("le_")
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"green", "yellow", "le_50_34", "le_80_78", "le_100", "le_inf"}
	h.SetLabels([]string{"green", "yellow"})
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}
}

func TestFVHistogram_Snapshot(t *testing.T) {
	h := NewFVHistogram([]float64{10, 20, 50, 80, 100}, nil)
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{0, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Labels(), got.Labels()) {
		t.Errorf("h.Snapshot().Labels() = %v, want %v", got.Labels(), h.Labels())
	}
	if got.NameTotal() != h.NameTotal() {
		t.Errorf("NewFFixedHistogram() total = %q, want %q", got.NameTotal(), h.NameTotal())
	}
	if !reflect.DeepEqual(h.Weights(), got.Weights()) {
		t.Errorf("h.Snapshot().Weights() = %v, want %v", got.Weights(), h.Weights())
	}
	if !reflect.DeepEqual(got.WeightsAliases(), h.WeightsAliases()) {
		t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), h.WeightsAliases())
	}
}

func BenchmarkFFixedHistogram(b *testing.B) {
	h := NewFFixedHistogram(10, 100, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFVHistogram05(b *testing.B) {
	h := NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFVHistogram20(b *testing.B) {
	h := NewFVHistogram(
		[]float64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFVHistogram100(b *testing.B) {
	h := NewFVHistogram(
		[]float64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFFixedHistogram_Values(b *testing.B) {
	h := NewFFixedHistogram(10, 100, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkFVHistogram05_Values(b *testing.B) {
	h := NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkFVHistogram20_Values(b *testing.B) {
	h := NewFVHistogram(
		[]float64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkFVHistogram100_Values(b *testing.B) {
	h := NewFVHistogram(
		[]float64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}
