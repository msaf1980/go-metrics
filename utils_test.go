package metrics

import (
	"math"
	"strconv"
	"testing"
)

func Test_searchInt64Ge(t *testing.T) {
	tests := []struct {
		a    []int64
		v    int64
		want int
	}{
		{a: []int64{math.MaxInt64}, v: 10},
		{a: []int64{math.MaxInt64}},
		{a: []int64{1, math.MaxInt64}, v: 0},
		{a: []int64{1, math.MaxInt64}, v: 1},
		{a: []int64{1, math.MaxInt64}, v: 2, want: 1},
		{a: []int64{1, math.MaxInt64}, v: 2, want: 1},
		{a: []int64{1, 2, 9, math.MaxInt64}},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 1, want: 0},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 2, want: 1},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 8, want: 2},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 9, want: 2},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 10, want: 3},
		{a: []int64{1, 2, 9, math.MaxInt64}, v: 11, want: 3},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 0},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 1, want: 0},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 2, want: 1},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 8, want: 2},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 9, want: 2},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 10, want: 3},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: 11, want: 4},
		{a: []int64{1, 2, 9, 10, math.MaxInt64}, v: -1},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := searchInt64Ge(tt.a, tt.v); got != tt.want {
				t.Errorf("searchUint64Ge(%+v, %d) = %v, want %v", tt.a, tt.v, got, tt.want)
			}
		})
	}
}

func Test_searchUint64Ge(t *testing.T) {
	tests := []struct {
		a    []uint64
		v    uint64
		want int
	}{
		{a: []uint64{math.MaxUint64}, v: 10},
		{a: []uint64{math.MaxUint64}},
		{a: []uint64{1, math.MaxUint64}, v: 0},
		{a: []uint64{1, math.MaxUint64}, v: 1},
		{a: []uint64{1, math.MaxUint64}, v: 2, want: 1},
		{a: []uint64{1, math.MaxUint64}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, math.MaxUint64}},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 1, want: 0},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 8, want: 2},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 9, want: 2},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 10, want: 3},
		{a: []uint64{1, 2, 9, math.MaxUint64}, v: 11, want: 3},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 1, want: 0},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 8, want: 2},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 9, want: 2},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 10, want: 3},
		{a: []uint64{1, 2, 9, 10, math.MaxUint64}, v: 11, want: 4},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := searchUint64Ge(tt.a, tt.v); got != tt.want {
				t.Errorf("searchUint64Ge(%+v, %d) = %v, want %v", tt.a, tt.v, got, tt.want)
			}
		})
	}
}

func Test_searchFloat64Ge(t *testing.T) {
	tests := []struct {
		a    []float64
		v    float64
		want int
	}{
		{a: []float64{math.MaxFloat64}, v: 10},
		{a: []float64{math.MaxFloat64}},
		{a: []float64{1, math.MaxFloat64}, v: 0},
		{a: []float64{1, math.MaxFloat64}, v: 1},
		{a: []float64{1, math.MaxFloat64}, v: 2, want: 1},
		{a: []float64{1, math.MaxFloat64}, v: 2, want: 1},
		{a: []float64{1, 2, 9, math.MaxFloat64}},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 1, want: 0},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 2, want: 1},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 8, want: 2},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 9, want: 2},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 10, want: 3},
		{a: []float64{1, 2, 9, math.MaxFloat64}, v: 11, want: 3},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 0},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 1, want: 0},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 2, want: 1},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 8, want: 2},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 9, want: 2},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 10, want: 3},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: 11, want: 4},
		{a: []float64{1, 2, 9, 10, math.MaxFloat64}, v: -1},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := searchFloat64Ge(tt.a, tt.v); got != tt.want {
				t.Errorf("searchFloat64Ge(%+v, %v) = %v, want %v", tt.a, tt.v, got, tt.want)
			}
		})
	}
}
