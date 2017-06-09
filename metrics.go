package main

import "github.com/prometheus/client_golang/prometheus"

var durationMetric = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "gateway_request_duration_milliseconds",
		Help:    "Request duration distribution",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.1, 140),
	})
