package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPLogger logs HTTP requests and collects request related metrics
type HTTPLogger struct {
	handler        http.Handler
	durationMetric prometheus.Histogram
}

func (l HTTPLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	l.handler.ServeHTTP(w, r)

	duration := time.Now().Sub(start)
	l.durationMetric.Observe(float64(duration) / float64(time.Millisecond))
}
