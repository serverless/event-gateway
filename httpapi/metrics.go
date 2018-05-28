package httpapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricEventTypes)
	prometheus.MustRegister(metricFunctions)
	prometheus.MustRegister(metricSubscriptions)

	prometheus.MustRegister(metricConfigRequests)
	prometheus.MustRegister(metricConfigRequestDuration)
}

// Event Types

var metricEventTypes = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "gateway",
		Subsystem: "eventtypes",
		Name:      "total",
		Help:      "Gauge of created event types count.",
	}, []string{"space"})

// Functions

var metricFunctions = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "gateway",
		Subsystem: "functions",
		Name:      "total",
		Help:      "Gauge of registered functions count.",
	}, []string{"space"})

// Subscriptions

var metricSubscriptions = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "gateway",
		Subsystem: "subscriptions",
		Name:      "total",
		Help:      "Gauge of created subscriptions count.",
	}, []string{"space"})

// Config API

var metricConfigRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "config",
		Name:      "requests_total",
		Help:      "Total of Config API requests.",
	}, []string{"space", "resource", "operation"})

var metricConfigRequestDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "gateway",
		Subsystem: "config",
		Name:      "request_duration_seconds",
		Help:      "Bucketed histogram of request duration of Config API requests",
		Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 16),
	})
