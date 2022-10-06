package test

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
)

func FloatEq(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}

func SliceFloatEq(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !FloatEq(a[i], b[i]) {
			return false
		}
	}
	return true
}

func FloatEqDev(a, b, dev float64) bool {
	return math.Abs(a-b) <= dev
}

type Value struct {
	V   float64
	Dev float64
}

func CompareMetrics(t *testing.T, expected map[string]Value, actual map[string]float64) bool {
	var errs []string
	for k, ev := range expected {
		if av, ok := actual[k]; !ok {
			errs = append(errs, fmt.Sprintf("\n- %q: %f", k, ev.V))
		} else if !FloatEqDev(av, ev.V, ev.Dev) {
			errs = append(errs, fmt.Sprintf("\n  %q: %f, want %f", k, av, ev.V))
		}
	}
	for k, av := range actual {
		if _, ok := expected[k]; !ok {
			errs = append(errs, fmt.Sprintf("\n+ %q: %f", k, av))
		}
	}
	if len(errs) != 0 {
		sort.Strings(errs)
		t.Error(strings.Join(errs, ""))
		return true
	}
	return false
}
