package metrics

import (
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
		prefix             string
		total              string
		wantWeights        []float64
		wantWeightsAliases []string
		wantNames          []string
	}{
		{
			startVal:           100,
			endVal:             1000,
			width:              100,
			wantWeights:        []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100},
			wantWeightsAliases: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000", "inf"},
			// wantNames:   []string{"0100", "0200", "0300", "0400", "0500", "0600", "0700", "0800", "0900", "1000", "inf"},
			wantNames: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000", "inf"},
		},
		{
			startVal:           10,
			endVal:             100,
			width:              40,
			prefix:             "req_le_",
			wantWeights:        []float64{10, 50, 90, 130, 170},
			wantWeightsAliases: []string{"10", "50", "90", "130", "inf"},
			wantNames:          []string{"req_le_10", "req_le_50", "req_le_90", "req_le_130", "req_le_inf"},
		},
		{
			startVal:           10.1,
			endVal:             100,
			width:              40.01,
			prefix:             "req_le_",
			wantWeights:        []float64{10.1, 50.11, 90.12, 130.13, 170.14},
			wantWeightsAliases: []string{"10_1", "50_11", "90_12", "130_13", "inf"},
			wantNames:          []string{"req_le_10_1", "req_le_50_11", "req_le_90_12", "req_le_130_13", "req_le_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewFFixedHistogram(tt.startVal, tt.endVal, tt.width, tt.prefix, tt.total)
			if !test.SliceFloatEq(got.Weights(), tt.wantWeights) {
				t.Errorf("NewUFFistogram() weights = %+v, want %+v", got.Weights(), tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.Names(), tt.wantNames) {
				t.Errorf("NewUFFistogram() names =\n%q\nwant\n%q", got.Names(), tt.wantNames)
			}
			if tt.total == "" {
				tt.total = "total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewUFFistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.Names()) != len(got.Values()) {
				t.Errorf("NewUFFistogram() buckets count =%d, want %d", len(got.Names()), len(got.Values()))
			}
		})
	}
}

func TestFFixedHistogram_Add(t *testing.T) {
	startVal := 10.0
	width := 10.0
	endVal := 50.
	h := NewFFixedHistogram(startVal, endVal, width, "", "")
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

func TestFFixedHistogram_Snapshot(t *testing.T) {
	startVal := 10.0
	width := 10.0
	endVal := 50.0
	r := NewRegistry()
	h := GetOrRegisterFFixedHistogram("histogram", r, startVal, endVal, width, "le_", "req_total")
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{0, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Names(), got.Names()) {
		t.Errorf("h.Snapshot().Names() = %v, want %v", got.Names(), h.Names())
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
		prefix             string
		total              string
		wantWeights        []float64
		wantWeightsAliases []string
		wantNames          []string
	}{
		{
			weights:            []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20},
			wantWeights:        []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20, 21},
			wantWeightsAliases: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "20", "inf"},
			// wantNames:   []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "20", "inf"},
			wantNames: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "20", "inf"},
		},
		{
			weights:            []float64{10, 20, 100},
			names:              []string{"green", "blue", "yellow", "red", "none"},
			wantWeights:        []float64{10, 20, 100, 101},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantNames:          []string{"green", "blue", "yellow", "red"},
		},
		{
			weights:            []float64{10, 20, 100},
			names:              []string{"green", "blue", "yellow"},
			prefix:             "req_",
			total:              "total_req",
			wantWeights:        []float64{10, 20, 100, 101},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantNames:          []string{"req_green", "req_blue", "req_yellow", "req_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewFVHistogram(tt.weights, tt.names, tt.prefix, tt.total)
			if !reflect.DeepEqual(got.weights, tt.wantWeights) {
				t.Errorf("NewFVHistogram() weights = %+v, want %+v", got.weights, tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.names, tt.wantNames) {
				t.Errorf("NewFVHistogram() names =\n%q\nwant\n%q", got.names, tt.wantNames)
			}
			if tt.total == "" {
				tt.total = "total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewFVHistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.names) != len(got.Values()) {
				t.Errorf("NewFVHistogram() buckets count =%d, want %d", len(got.Values()), len(tt.wantNames))
			}
		})
	}
}

func TestFVHistogram_Add(t *testing.T) {
	r := NewRegistry()
	h := GetOrRegisterFVHistogram("histogram", r, []float64{1, 2, 5, 8, 20}, nil, "", "")
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

func TestFVHistogram_Snapshot(t *testing.T) {
	h := NewFVHistogram([]float64{10, 20, 50, 80, 100}, nil, "", "")
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{0, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Names(), got.Names()) {
		t.Errorf("h.Snapshot().Names() = %v, want %v", got.Names(), h.Names())
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
	h := NewFFixedHistogram(10, 100, 10, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFVHistogram05(b *testing.B) {
	h := NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFVHistogram20(b *testing.B) {
	h := NewFVHistogram(
		[]float64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil, "", "")
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
		}, nil, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFFixedHistogram_Values(b *testing.B) {
	h := NewFFixedHistogram(10, 100, 10, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkFVHistogram05_Values(b *testing.B) {
	h := NewFVHistogram([]float64{10, 50, 100, 200, 300}, nil, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkFVHistogram20_Values(b *testing.B) {
	h := NewFVHistogram(
		[]float64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil, "", "")
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
		}, nil, "", "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}