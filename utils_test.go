package metrics

import (
	"strconv"
	"testing"
)

func Test_searchUint64Ge(t *testing.T) {
	tests := []struct {
		a    []uint64
		v    uint64
		want int
	}{
		{a: []uint64{1}, v: 10},
		{a: []uint64{1}},
		{a: []uint64{1, 10}, v: 0},
		{a: []uint64{1, 10}, v: 1},
		{a: []uint64{1, 10}, v: 2, want: 1},
		{a: []uint64{1, 10}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, 10}},
		{a: []uint64{1, 2, 9, 10}, v: 1, want: 0},
		{a: []uint64{1, 2, 9, 10}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, 10}, v: 8, want: 2},
		{a: []uint64{1, 2, 9, 10}, v: 9, want: 2},
		{a: []uint64{1, 2, 9, 10}, v: 10, want: 3},
		{a: []uint64{1, 2, 9, 10}, v: 11, want: 3},
		{a: []uint64{1, 2, 9, 10, 11}, v: 1, want: 0},
		{a: []uint64{1, 2, 9, 10, 11}, v: 2, want: 1},
		{a: []uint64{1, 2, 9, 10, 11}, v: 8, want: 2},
		{a: []uint64{1, 2, 9, 10, 11}, v: 9, want: 2},
		{a: []uint64{1, 2, 9, 10, 11}, v: 10, want: 3},
		{a: []uint64{1, 2, 9, 10, 11}, v: 11, want: 4},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := searchUint64Ge(tt.a, tt.v); got != tt.want {
				t.Errorf("searchUint64Ge(%+v, %d) = %v, want %v", tt.a, tt.v, got, tt.want)
			}
		})
	}
}
