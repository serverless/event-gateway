package httpapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricFunctions)
	prometheus.MustRegister(metricSubscriptions)

	prometheus.MustRegister(metricConfigRequests)
}

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
