package metrics

import "github.com/prometheus/client_golang/prometheus"

// DurationMetric is a histogram with buckets that
// are incrementally 10% larger than the last, with valid
// values ranging from 0.1 to ~62370
var DurationMetric = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "gateway_request_duration_milliseconds",
		Help:    "Request duration distribution",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.1, 140),
	})
