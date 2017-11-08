package api

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(requestDuration)
}

var requestDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "gateway",
		Subsystem: "config",
		Name:      "request_duration_seconds",
		Help:      "Bucketed histogram of request duration of config API requests",
		Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 16),
	})

type metricsReporter struct {
	Handler http.Handler
}

func (m metricsReporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	m.Handler.ServeHTTP(w, r)
	requestDuration.Observe(time.Since(start).Seconds())
}
