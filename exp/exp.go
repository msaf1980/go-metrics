// Hook go-metrics into expvar
// on any /debug/metrics request, load all vars from the registry into expvar, and execute regular expvar handler
package exp

import (
	"fmt"
	"log"
	"net/http"

	"github.com/msaf1980/go-metrics"
)

type exp struct {
	registry metrics.Registry
	minLock  bool
}

func (exp *exp) expHandler(w http.ResponseWriter, r *http.Request) {
	// load our variables into expvar
	// now just run the official expvar handler code (which is not publicly callable, so pasted inline)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{")
	first := true
	exp.registry.Each(func(name, tags string, tagsMap map[string]string, i interface{}) error {
		if first {
			first = false
		} else {
			fmt.Fprint(w, ",")
		}
		switch metric := i.(type) {
		case metrics.Counter:
			fmt.Fprintf(w, "\n  \"%s%s\": %d", name, tags, metric.Count())
		case metrics.DownCounter:
			fmt.Fprintf(w, "\n  \"%s%s\": %d", name, tags, metric.Count())
		case metrics.Gauge:
			fmt.Fprintf(w, "\n  \"%s%s\": %d", name, tags, metric.Value())
		case metrics.GaugeFloat64:
			fmt.Fprintf(w, "\n  \"%s%s\": %f", name, tags, metric.Value())
		case metrics.Healthcheck:
			fmt.Fprintf(w, "\n  \"%s%s\": %d", name, tags, metric.Check())
		case metrics.HistogramInterface:
			vals := metric.Values()
			leAliases := metric.WeightsAliases()
			if metric.IsSummed() {
				for i, label := range metric.Labels() {
					if i > 0 {
						fmt.Fprint(w, ",")
					}
					if tags == "" {
						fmt.Fprintf(w, "\n  \"%s%s%s\": %d", name, label, tags, vals[i])
					} else {
						fmt.Fprintf(w, "\n  \"%s%s%s\": %d", name, label, tags+";le="+leAliases[i], vals[i])
					}
				}
				if tags == "" {
					fmt.Fprintf(w, ",\n  \"%s%s%s\": %d", name, metric.NameTotal(), tags, vals[0])
				} else {
					fmt.Fprintf(w, ",\n  \"%s%s%s\": %d", name, metric.NameTotal(), tags, vals[0])
				}
			} else {
				var total uint64
				for i, label := range metric.Labels() {
					if i > 0 {
						fmt.Fprint(w, ",")
					}
					if tags == "" {
						fmt.Fprintf(w, "\n  \"%s%s%s\": %d", name, label, tags, vals[i])
					} else {
						fmt.Fprintf(w, "\n  \"%s%s%s\": %d", name, label, tags+";le="+leAliases[i], vals[i])
					}
					total += vals[i]
				}
				if tags == "" {
					fmt.Fprintf(w, ",\n  \"%s%s%s\": %d", name, metric.NameTotal(), tags, total)
				} else {
					fmt.Fprintf(w, ",\n  \"%s%s%s\": %d", name, metric.NameTotal(), tags, total)
				}
			}
		case metrics.Rate:
			v, rate := metric.Values()
			fmt.Fprintf(w, "\n  \"%s%s%s\": %d,", name, metric.Name(), tags, v)
			fmt.Fprintf(w, "\n  \"%s%s%s\": %f", name, metric.RateName(), tags, rate)
		case metrics.FRate:
			v, rate := metric.Values()
			fmt.Fprintf(w, "\n  \"%s%s%s\": %f,", name, metric.Name(), tags, v)
			fmt.Fprintf(w, "\n  \"%s%s%s\": %f", name, metric.RateName(), tags, rate)
		default:
			fmt.Fprintf(w, "\n  \"%s%s\": NaN", name, tags)
			log.Printf("\n  \"%s%s\": \"<UHHADLED:%T>\"", name, tags, i)
		}
		return nil
	}, exp.minLock)
	if first {
		fmt.Fprintf(w, "}\n")
	} else {
		fmt.Fprintf(w, "\n}\n")
	}
}

// Exp will register an expvar powered metrics handler with http.DefaultServeMux on "/debug/vars"
func Exp(r metrics.Registry, minLock bool) {
	h := ExpHandler(r, minLock)
	// this would cause a panic:
	// panic: http: multiple registrations for /debug/vars
	// http.HandleFunc("/debug/vars", e.expHandler)
	// haven't found an elegant way, so just use a different endpoint
	http.Handle("/debug/metrics", h)
}

// ExpHandler will return an expvar powered metrics handler.
func ExpHandler(r metrics.Registry, minLock bool) http.Handler {
	e := exp{registry: r, minLock: minLock}
	return http.HandlerFunc(e.expHandler)
}
