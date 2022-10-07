package metrics

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
)

func BenchmarkRegistryMinLock(b *testing.B) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, string, map[string]string, interface{}) error { return nil }, false)
	}
}

func BenchmarkRegistry(b *testing.B) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, string, map[string]string, interface{}) error { return nil }, true)
	}
}

func BenchmarkRegistryTMinLock(b *testing.B) {
	r := NewRegistry()
	r.RegisterT("foo", map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, string, map[string]string, interface{}) error { return nil }, false)
	}
}

func BenchmarkRegistryT(b *testing.B) {
	r := NewRegistry()
	r.RegisterT("foo", map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Each(func(string, string, map[string]string, interface{}) error { return nil }, true)
	}
}

func BenchmarkRegistry1000MinLock(b *testing.B) {
	r := NewRegistry()
	n := 1000
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo%07d", i), NewCounter())
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, true)
	}
}

func BenchmarkRegistry1000(b *testing.B) {
	r := NewRegistry()
	n := 1000
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo%07d", i), NewCounter())
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, false)
	}
}

func BenchmarkRegistry1000TMinLock(b *testing.B) {
	r := NewRegistry()
	n := 1000
	for i := 0; i < n; i++ {
		r.RegisterT(fmt.Sprintf("foo%07d", i), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, true)
	}
}

func BenchmarkRegistry1000T(b *testing.B) {
	r := NewRegistry()
	n := 1000
	for i := 0; i < n; i++ {
		r.RegisterT(fmt.Sprintf("foo%07d", i), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, false)
	}
}

func BenchmarkRegistry10000MinLock(b *testing.B) {
	r := NewRegistry()
	n := 10000
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo%07d", i), NewCounter())
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, true)
	}
}

func BenchmarkRegistry10000(b *testing.B) {
	r := NewRegistry()
	n := 10000
	for i := 0; i < n; i++ {
		r.Register(fmt.Sprintf("foo%07d", i), NewCounter())
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, false)
	}
}

func BenchmarkRegistry10000TMinLock(b *testing.B) {
	r := NewRegistry()
	n := 10000
	for i := 0; i < n; i++ {
		r.RegisterT(fmt.Sprintf("foo%07d", i), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, true)
	}
}

func BenchmarkRegistry10000T(b *testing.B) {
	r := NewRegistry()
	n := 10000
	for i := 0; i < n; i++ {
		r.RegisterT(fmt.Sprintf("foo%07d", i), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	}
	v := make([]string, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := v[:0]
		r.Each(func(k, _ string, _ map[string]string, _ interface{}) error {
			v = append(v, k)
			return nil
		}, false)
	}
}

func BenchmarkRegistry10000_Register(b *testing.B) {
	r := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetOrRegister(strconv.FormatInt(int64(i), 10), NewCounter)
	}
}

func BenchmarkRegistry10000_RegisterT(b *testing.B) {
	r := NewRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetOrRegisterT(strconv.FormatInt(int64(i), 10), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
	}
}

func BenchmarkRegistryParallel(b *testing.B) {
	var i int64
	r := NewRegistry()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetOrRegister(strconv.FormatInt(atomic.AddInt64(&i, 1), 10), NewCounter)
		}
	})
}

func BenchmarkRegistryParallelT(b *testing.B) {
	var i int64
	r := NewRegistry()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetOrRegisterT(strconv.FormatInt(atomic.AddInt64(&i, 1), 10), map[string]string{"tag1": "value1", "tag21": "value21"}, NewCounter)
		}
	})
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	i := 0
	r.Each(func(name, _ string, _ map[string]string, iface interface{}) error {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, true)
	if i != 1 {
		t.Fatal(i)
	}
	r.Unregister("foo")
	i = 0
	r.Each(func(string, string, map[string]string, interface{}) error { i++; return nil }, false)
	if i != 0 {
		t.Fatal(i)
	}
}

func TestRegistryDuplicate(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("foo", NewCounter()); nil != err {
		t.Fatal(err)
	}
	if err := r.Register("foo", NewGauge()); nil == err {
		t.Fatal(err)
	}
	i := 0
	r.Each(func(name, tags string, _ map[string]string, iface interface{}) error {
		i++
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, true)
	if i != 1 {
		t.Fatal(i)
	}
}

func TestGetOrRegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	c := GetOrRegisterCounter("foo", r)
	if c == nil {
		t.Fatal("not a counter")
	}
	c2 := GetOrRegisterCounter("foo", r)
	if c != c2 {
		t.Fatal("duplicate")
	}
	i := 0
	r.Each(func(name, tags string, _ map[string]string, iface interface{}) error {
		i++
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, true)
	if i != 1 {
		t.Fatal(i)
	}
}

func TestGetOrRegisterTypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("must panic")
		}
	}()
	r := NewRegistry()
	var s string
	_ = r.GetOrRegister("foo", s)
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	r.Register("foo", NewCounter())
	if count := r.Get("foo").(Counter).Count(); count != 0 {
		t.Fatal(count)
	}
	r.Get("foo").(Counter).Inc(1)
	if count := r.Get("foo").(Counter).Count(); count != 1 {
		t.Fatal(count)
	}
}

func TestRegistryGetOrRegister(t *testing.T) {
	r := NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", NewCounter())
	m := r.GetOrRegister("foo", NewGauge())
	if _, ok := m.(Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name, tags string, _ map[string]string, iface interface{}) error {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, false)
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryGetOrRegisterMinLock(t *testing.T) {
	r := NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", NewCounter())
	m := r.GetOrRegister("foo", NewGauge())
	if _, ok := m.(Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name, tags string, _ map[string]string, iface interface{}) error {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, true)
	if i != 1 {
		t.Fatal(i)
	}
}

func TestRegistryGetOrRegisterWithLazyInstantiation(t *testing.T) {
	r := NewRegistry()

	// First metric wins with GetOrRegister
	_ = r.GetOrRegister("foo", NewCounter)
	m := r.GetOrRegister("foo", NewGauge)
	if _, ok := m.(Counter); !ok {
		t.Fatal(m)
	}

	i := 0
	r.Each(func(name, tags string, _ map[string]string, iface interface{}) error {
		i++
		if name != "foo" {
			t.Fatal(name)
		}
		if _, ok := iface.(Counter); !ok {
			t.Fatal(iface)
		}
		return nil
	}, false)
	if i != 1 {
		t.Fatal(i)
	}
}

// func TestRegistryUnregister(t *testing.T) {
// 	l := len(arbiter.meters)
// 	r := NewRegistry()
// 	r.Register("foo", NewCounter())
// 	r.Register("bar", NewMeter())
// 	r.Register("baz", NewTimer())
// 	if len(arbiter.meters) != l+2 {
// 		t.Errorf("arbiter.meters: %d != %d\n", l+2, len(arbiter.meters))
// 	}
// 	r.Unregister("foo")
// 	r.Unregister("bar")
// 	r.Unregister("baz")
// 	if len(arbiter.meters) != l {
// 		t.Errorf("arbiter.meters: %d != %d\n", l+2, len(arbiter.meters))
// 	}
// }
