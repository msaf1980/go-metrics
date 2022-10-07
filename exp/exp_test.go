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

	h := metrics.NewUFixedHistogram(1, 3, 1).AddLabelPrefix("req_")
	h.Add(2)
	if err := r.Register("histogram", h); err != nil {
		t.Fatal(err)
	}

	sh := metrics.NewFixedSumHistogram(1, 3, 1).AddLabelPrefix("req_")
	sh.Add(2)
	h.Add(6)
	if err := r.Register("shistogram", sh); err != nil {
		t.Fatal(err)
	}

	ht := metrics.NewVHistogram([]int64{1, 2, 5, 8, 20}, nil).AddLabelPrefix("req_")
	ht.Add(2)
	ht.Add(6)
	if err := r.RegisterT("histogram", map[string]string{"tag1": "value1", "tag21": "value21"}, ht); err != nil {
		t.Fatal(err)
	}

	rate := metrics.GetOrRegisterRate("ratefoo", r).SetName("_value").SetRateName("_rate")
	rate.UpdateTs(1, 1e9)
	rate.UpdateTs(7, 3e9)

	rate2 := metrics.GetOrRegisterRate("ratefoo2", r)
	rate2.UpdateTs(2, 1e9)
	rate2.UpdateTs(8, 3e9)

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
		"histogram.req_inf":               1,
		"histogram.total":                 2,
		"histogram.req_1;tag1=value1;tag21=value21;le=1":     0,
		"histogram.req_2;tag1=value1;tag21=value21;le=2":     1,
		"histogram.req_5;tag1=value1;tag21=value21;le=5":     0,
		"histogram.req_8;tag1=value1;tag21=value21;le=8":     1,
		"histogram.req_20;tag1=value1;tag21=value21;le=20":   0,
		"histogram.req_inf;tag1=value1;tag21=value21;le=inf": 0,
		"histogram.total;tag1=value1;tag21=value21":          2,
		"shistogram.req_1":   1,
		"shistogram.req_2":   1,
		"shistogram.req_3":   0,
		"shistogram.req_inf": 0,
		"shistogram.total":   1,
		"ratefoo_value":      7,
		"ratefoo_rate":       3,
		"ratefoo2.value":     8,
		"ratefoo2.rate":      3,
	}
	if err = json.Unmarshal(body, &got); err == nil {
		assert.Equal(t, want, got)
	} else {
		t.Fatal(err, "\n", string(body))
	}
}
