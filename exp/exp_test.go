// Hook go-metrics into expvar
// on any /debug/metrics request, load all vars from the registry into expvar, and execute regular expvar handler
package exp

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/msaf1980/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/xhhuango/json"
)

const keyAddress = "address"

func TestExp(t *testing.T) {
	r := metrics.NewRegistry()

	c := metrics.GetOrRegisterCounterT("count", map[string]string{"tag1": "value1", "tag21": "value21"}, r)
	c.Inc(46)

	dc := metrics.GetOrRegisterDownCounter("count", r)
	dc.Dec(4)

	h := metrics.NewUFixedHistogram(1, 3, 1, "req_", "")
	h.Add(2)
	if err := r.Register("histogram", h); err != nil {
		t.Fatal(err)
	}

	ht := metrics.NewVHistogram([]int64{1, 2, 5, 8, 20}, nil, "req_", "")
	ht.Add(2)
	ht.Add(6)
	if err := r.RegisterT("histogram", map[string]string{"tag1": "value1", "tag21": "value21"}, ht); err != nil {
		t.Fatal(err)
	}

	rate := metrics.GetOrRegisterRate("ratefoo", r)
	rate.Update(1, 1e9)
	rate.Update(7, 3e9)

	mux := http.NewServeMux()
	mux.Handle("/debug/metrics", ExpHandler(r, false))

	lsnr, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := &http.Server{
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.WithValue(ctx, keyAddress, l.Addr().String())
			return ctx
		},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.Serve(lsnr)
	}()
	time.Sleep(time.Millisecond)

	req, err := http.Get("http://" + lsnr.Addr().String() + "/debug/metrics")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if req.StatusCode != http.StatusOK {
		t.Fatal(req.Status, string(body))
	}
	var got map[string]float64
	want := map[string]float64{
		"count;tag1=value1;tag21=value21": 46,
		"count":                           -4,
		"histogram.req_1":                 0,
		"histogram.req_2":                 1,
		"histogram.req_3":                 0,
		"histogram.req_inf":               0,
		"histogram.total":                 1,
		"histogram;tag1=value1;tag21=value21;label=req_1;le=1":     0,
		"histogram;tag1=value1;tag21=value21;label=req_2;le=2":     1,
		"histogram;tag1=value1;tag21=value21;label=req_5;le=5":     0,
		"histogram;tag1=value1;tag21=value21;label=req_8;le=8":     1,
		"histogram;tag1=value1;tag21=value21;label=req_20;le=20":   0,
		"histogram;tag1=value1;tag21=value21;label=req_inf;le=inf": 0,
		"histogram;tag1=value1;tag21=value21;label=total":          2,
		"ratefoo.value": 6,
		"ratefoo.rate":  3,
	}
	if err = json.Unmarshal(body, &got); err == nil {
		assert.Equal(t, want, got)
	} else {
		t.Fatal(err, "\n", string(body))
	}
}
