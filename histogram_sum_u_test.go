package metrics

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestNewFixedSumUHistogram(t *testing.T) {
	tests := []struct {
		startVal           uint64
		endVal             uint64
		width              uint64
		labelPrefix        string
		total              string
		wantWeights        []uint64
		wantWeightsAliases []string
		wantLabels         []string
	}{
		{
			startVal:           100,
			endVal:             1000,
			width:              100,
			wantWeights:        []uint64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100},
			wantWeightsAliases: []string{"100", "200", "300", "400", "500", "600", "700", "800", "900", "1000", "inf"},
			// wantLabels:   []string{"0100", "0200", "0300", "0400", "0500", "0600", "0700", "0800", "0900", "1000", "inf"},
			wantLabels: []string{".100", ".200", ".300", ".400", ".500", ".600", ".700", ".800", ".900", ".1000", ".inf"},
		},
		{
			startVal:           10,
			endVal:             100,
			width:              40,
			labelPrefix:        "req_le_",
			wantWeights:        []uint64{10, 50, 90, 130, 170},
			wantWeightsAliases: []string{"10", "50", "90", "130", "inf"},
			// wantLabels:   []string{"req_le_010", "req_le_050", "req_le_090", "req_le_130", "req_le_inf"},
			wantLabels: []string{".req_le_10", ".req_le_50", ".req_le_90", ".req_le_130", ".req_le_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewFixedSumUHistogram(tt.startVal, tt.endVal, tt.width)
			if tt.labelPrefix != "" {
				got.AddLabelPrefix(tt.labelPrefix)
			}
			if tt.total != "" {
				got.SetNameTotal(tt.total)
			}
			if !reflect.DeepEqual(got.Weights(), tt.wantWeights) {
				t.Errorf("NewFixedSumUHistogram() weights = %+v, want %+v", got.Weights(), tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.Labels(), tt.wantLabels) {
				t.Errorf("NewFixedSumUHistogram() names =\n%q\nwant\n%q", got.Labels(), tt.wantLabels)
			}
			if tt.total == "" {
				tt.total = ".total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewFixedSumUHistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.Labels()) != len(got.Values()) {
				t.Errorf("NewFixedSumUHistogram() buckets count =%d, want %d", len(got.Labels()), len(got.Values()))
			}
		})
	}
}

func TestFixedSumUHistogram_Add(t *testing.T) {
	startVal := uint64(10)
	width := uint64(10)
	endVal := uint64(50)
	r := NewRegistry()
	h := GetOrRegisterFixedSumUHistogram("histogram", r, startVal, endVal, width)
	// 10 20 30 40 50 inf
	tests := []struct {
		add  uint64
		want []uint64
	}{
		{add: 0, want: []uint64{1, 0, 0, 0, 0, 0}},
		{add: 10, want: []uint64{2, 0, 0, 0, 0, 0}},
		{add: 11, want: []uint64{3, 1, 0, 0, 0, 0}},
		{add: 20, want: []uint64{4, 2, 0, 0, 0, 0}},
		{add: 25, want: []uint64{5, 3, 1, 0, 0, 0}},
		{add: 49, want: []uint64{6, 4, 2, 1, 1, 0}},
		{add: 50, want: []uint64{7, 5, 3, 2, 2, 0}},
		{add: 51, want: []uint64{8, 6, 4, 3, 3, 1}},
		{add: 100, want: []uint64{9, 7, 5, 4, 4, 2}},
	}

	// zero
	got := h.Values()
	if !reflect.DeepEqual([]uint64{0, 0, 0, 0, 0, 0}, got) {
		t.Errorf("h.Values() = %v, want zero %v", got, []uint64{0, 0, 0, 0, 0, 0})
	}

	for n, tt := range tests {
		t.Run(strconv.FormatUint(tt.add, 10)+"#"+strconv.Itoa(n), func(t *testing.T) {
			h.Add(tt.add)
			got := h.Values()
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("h.Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixedSumUHistogram_SetNames(t *testing.T) {
	startVal := uint64(10)
	width := uint64(10)
	endVal := uint64(50)
	h := NewFixedSumUHistogram(startVal, endVal, width)

	wantLabels := []string{".10", ".20", ".30", ".40", ".50", ".inf"}
	weightsAliases := []string{"10", "20", "30", "40", "50", "inf"}
	wantNameTotal := ".total"
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{".le_10", ".le_20", ".le_30", ".le_40", ".le_50", ".le_inf"}
	wantNameTotal = "req_total"
	h.AddLabelPrefix("le_")
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"green", "yellow", ".le_30", ".le_40", ".le_50", ".le_inf"}
	h.SetLabels([]string{"green", "yellow"})
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}
}

func TestFixedSumUHistogram_Snapshot(t *testing.T) {
	startVal := uint64(10)
	width := uint64(10)
	endVal := uint64(50)
	h := NewFixedSumUHistogram(startVal, endVal, width)
	h.AddLabelPrefix("le_")
	h.SetNameTotal("req_total")
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{1, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Labels(), got.Labels()) {
		t.Errorf("h.Snapshot().Labels() = %v, want %v", got.Labels(), h.Labels())
	}
	if got.NameTotal() != h.NameTotal() {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", got.NameTotal(), h.NameTotal())
	}
	if !reflect.DeepEqual(h.Weights(), got.Weights()) {
		t.Errorf("h.Snapshot().Weights() = %v, want %v", got.Weights(), h.Weights())
	}
	if !reflect.DeepEqual(got.WeightsAliases(), h.WeightsAliases()) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), h.WeightsAliases())
	}
}

func TestInvalidVSumUHistogram(t *testing.T) {
	tests := []struct {
		weights []uint64
	}{
		{weights: []uint64{1, 1, 2, 3}},
		{weights: []uint64{1, 2, 2, 3}},
		{weights: []uint64{1, 2, 3, 3}},
		{weights: []uint64{1, 0, 2, 3}},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("must panic")
				}
			}()
			_ = NewVSumUHistogram(tt.weights, nil)
		})
	}
}

func TestNewVSumUHistogram(t *testing.T) {
	tests := []struct {
		weights            []uint64
		names              []string
		labelPrefix        string
		total              string
		wantWeights        []uint64
		wantWeightsAliases []string
		wantLabels         []string
	}{
		{
			weights:            []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20},
			wantWeights:        []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 20, math.MaxUint64},
			wantWeightsAliases: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "20", "inf"},
			// wantLabels:   []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "20", "inf"},
			wantLabels: []string{".1", ".2", ".3", ".4", ".5", ".6", ".7", ".8", ".9", ".20", ".inf"},
		},
		{
			weights:            []uint64{10, 20, 100},
			names:              []string{"green", "blue", "yellow", "red", "none"},
			wantWeights:        []uint64{10, 20, 100, math.MaxUint64},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantLabels:         []string{"green", "blue", "yellow", "red"},
		},
		{
			weights:            []uint64{10, 20, 100},
			names:              []string{"green", "blue", "yellow"},
			labelPrefix:        "req_",
			total:              "total_req",
			wantWeights:        []uint64{10, 20, 100, math.MaxUint64},
			wantWeightsAliases: []string{"10", "20", "100", "inf"},
			wantLabels:         []string{"req_green", "req_blue", "req_yellow", ".req_inf"},
		},
	}
	for i, tt := range tests {
		t.Run("#"+strconv.Itoa(i), func(t *testing.T) {
			got := NewVSumUHistogram(tt.weights, tt.names)
			if tt.labelPrefix != "" {
				got.AddLabelPrefix(tt.labelPrefix)
			}
			if tt.total != "" {
				got.SetNameTotal(tt.total)
			}
			if !reflect.DeepEqual(got.weights, tt.wantWeights) {
				t.Errorf("NewVSumUHistogram() weights = %+v, want %+v", got.weights, tt.wantWeights)
			}
			if !reflect.DeepEqual(got.WeightsAliases(), tt.wantWeightsAliases) {
				t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), tt.wantWeightsAliases)
			}
			if !reflect.DeepEqual(got.labels, tt.wantLabels) {
				t.Errorf("NewVSumUHistogram() names =\n%q\nwant\n%q", got.labels, tt.wantLabels)
			}
			if tt.total == "" {
				tt.total = ".total"
			}
			if got.NameTotal() != tt.total {
				t.Errorf("NewVSumUHistogram() total = %q, want %q", got.NameTotal(), tt.total)
			}
			if len(got.labels) != len(got.Values()) {
				t.Errorf("NewVSumUHistogram() buckets count =%d, want %d", len(got.Values()), len(tt.wantLabels))
			}
		})
	}
}

func TestVSumUHistogram_Add(t *testing.T) {
	r := NewRegistry()
	h := GetOrRegisterVSumUHistogram("histogram", r, []uint64{1, 2, 5, 8, 20}, nil)
	// 1, 2, 5, 8, 20, inf
	tests := []struct {
		add  uint64
		want []uint64
	}{
		{add: 0, want: []uint64{1, 0, 0, 0, 0, 0}},
		{add: 1, want: []uint64{2, 0, 0, 0, 0, 0}},
		{add: 2, want: []uint64{3, 1, 0, 0, 0, 0}},
		{add: 3, want: []uint64{4, 2, 1, 0, 0, 0}},
		{add: 5, want: []uint64{5, 3, 2, 0, 0, 0}},
		{add: 6, want: []uint64{6, 4, 3, 1, 0, 0}},
		{add: 7, want: []uint64{7, 5, 4, 2, 0, 0}},
		{add: 8, want: []uint64{8, 6, 5, 3, 0, 0}},
		{add: 9, want: []uint64{9, 7, 6, 4, 1, 0}},
		{add: 20, want: []uint64{10, 8, 7, 5, 2, 0}},
		{add: 21, want: []uint64{11, 9, 8, 6, 3, 1}},
		{add: 100, want: []uint64{12, 10, 9, 7, 4, 2}},
	}

	// zero
	got := h.Values()
	if !reflect.DeepEqual([]uint64{0, 0, 0, 0, 0, 0}, got) {
		t.Errorf("h.Values() = %v, want zero %v", got, []uint64{0, 0, 0, 0, 0, 0})
	}

	for n, tt := range tests {
		t.Run(strconv.FormatUint(tt.add, 10)+"#"+strconv.Itoa(n), func(t *testing.T) {
			h.Add(tt.add)
			got := h.Values()
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("h.Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVSumUHistogram_SetNames(t *testing.T) {
	h := NewVSumUHistogram([]uint64{10, 20, 50, 80, 100}, nil)

	wantLabels := []string{".10", ".20", ".50", ".80", ".100", ".inf"}
	weightsAliases := []string{"10", "20", "50", "80", "100", "inf"}
	wantNameTotal := ".total"
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{".le_10", ".le_20", ".le_50", ".le_80", ".le_100", ".le_inf"}
	wantNameTotal = "req_total"
	h.AddLabelPrefix("le_")
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}

	wantLabels = []string{"green", "yellow", ".le_50", ".le_80", ".le_100", ".le_inf"}
	h.SetLabels([]string{"green", "yellow"})
	h.SetNameTotal(wantNameTotal)
	if !reflect.DeepEqual(h.Labels(), wantLabels) {
		t.Errorf("h.Snapshot().Labels() = %q, want %q", h.Labels(), wantLabels)
	}
	if h.NameTotal() != wantNameTotal {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", h.NameTotal(), wantNameTotal)
	}
	if !reflect.DeepEqual(h.WeightsAliases(), weightsAliases) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", h.WeightsAliases(), weightsAliases)
	}
}

func TestVSumUHistogram_Snapshot(t *testing.T) {
	h := NewVSumUHistogram([]uint64{10, 20, 50, 80, 100}, nil)
	h.Add(19)
	got := h.Snapshot()
	want := []uint64{1, 1, 0, 0, 0, 0}
	if !reflect.DeepEqual(want, got.Values()) {
		t.Errorf("h.Snapshot().Values() = %v, want %v", got.Values(), want)
	}
	if !reflect.DeepEqual(h.Labels(), got.Labels()) {
		t.Errorf("h.Snapshot().Labels() = %v, want %v", got.Labels(), h.Labels())
	}
	if got.NameTotal() != h.NameTotal() {
		t.Errorf("NewFixedSumUHistogram() total = %q, want %q", got.NameTotal(), h.NameTotal())
	}
	if !reflect.DeepEqual(h.Weights(), got.Weights()) {
		t.Errorf("h.Snapshot().Weights() = %v, want %v", got.Weights(), h.Weights())
	}
	if !reflect.DeepEqual(got.WeightsAliases(), h.WeightsAliases()) {
		t.Errorf("NewFixedSumUHistogram() weightsAliases =\n%q\nwant\n%q", got.WeightsAliases(), h.WeightsAliases())
	}
}

func BenchmarkFixedSumUHistogram(b *testing.B) {
	h := NewFixedSumUHistogram(10, 100, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkFixedSumUHistogramH(b *testing.B) {
	h := NewFixedSumUHistogram(10, 100, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(1000)
	}
}

func BenchmarkFixedSumUHistogramParallel(b *testing.B) {
	h := NewFixedSumUHistogram(10, 100, 10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(50)
		}
	})
}

func BenchmarkFixedSumUHistogramParallelH(b *testing.B) {
	h := NewFixedSumUHistogram(10, 100, 10)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(1000)
		}
	})
}

func BenchmarkVSumUHistogram05(b *testing.B) {
	h := NewVSumUHistogram([]uint64{10, 50, 100, 200, 300}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkVSumUHistogram05Parallel(b *testing.B) {
	h := NewVSumUHistogram([]uint64{10, 50, 100, 200, 300}, nil)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(50)
		}
	})
}

func BenchmarkVSumUHistogram20(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
	}
}

func BenchmarkVSumUHistogram20Parallel(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(50)
		}
	})
}

func BenchmarkVSumUHistogram20ParallelH(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(10000)
		}
	})
}

func BenchmarkVSumUHistogram100(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{
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

func BenchmarkVSumUHistogram100H(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(10000)
	}
}

func BenchmarkVSumUHistogram100Parallel(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(50)
		}
	})
}

func BenchmarkVSumUHistogram100ParallelH(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Add(10000)
		}
	})
}

func BenchmarkFixedSumUHistogram_Values(b *testing.B) {
	h := NewFixedSumUHistogram(10, 100, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkVSumUHistogram05_Values(b *testing.B) {
	h := NewVSumUHistogram([]uint64{10, 50, 100, 200, 300}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(50)
		_ = h.Values()
	}
}

func BenchmarkVSumUHistogram20_Values(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800},
		nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(1)
		_ = h.Values()
	}
}

func BenchmarkVSumUHistogram100_Values(b *testing.B) {
	h := NewVSumUHistogram(
		[]uint64{
			10, 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800,
			1900, 2000, 2100, 2200, 2300, 2400, 2500, 2600, 2700, 2800, 2900, 3000, 3100, 3200, 3300, 3400, 3500, 3600, 3700, 3800,
			3900, 4000, 4100, 4200, 4300, 4400, 4500, 4600, 4700, 4800, 4900, 5000, 5100, 5200, 5300, 5400, 5500, 5600, 5700, 5800,
			5900, 6000, 6100, 6200, 6300, 6400, 6500, 6600, 6700, 6800, 6900, 7000, 7100, 7200, 7300, 7400, 7500, 7600, 7700, 7800,
			6900, 7000, 8100, 8200, 8300, 8400, 8500, 8600, 8700, 8800, 8900, 9000, 9100, 9200, 9300, 9400, 9500, 9600, 9700, 9800,
		}, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Add(1)
		_ = h.Values()
	}
}
