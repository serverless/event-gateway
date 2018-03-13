package httpapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricFunctionRegistered)
	prometheus.MustRegister(metricFunctionDeleted)
	prometheus.MustRegister(metricSubscriptionCreated)
	prometheus.MustRegister(metricSubscriptionDeleted)

	prometheus.MustRegister(metricConfigRequests)
}

// Functions

var metricFunctionRegistered = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "functions",
		Name:      "registered_total",
		Help:      "Total of functions registered.",
	}, []string{"space"})

var metricFunctionDeleted = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "functions",
		Name:      "deleted_total",
		Help:      "Total of functions deleted.",
	}, []string{"space"})

// Subscriptions

var metricSubscriptionCreated = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "subscriptions",
		Name:      "created_total",
		Help:      "Total of subscriptions created.",
	}, []string{"space"})

var metricSubscriptionDeleted = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "subscriptions",
		Name:      "deleted_total",
		Help:      "Total of subscriptions deleted.",
	}, []string{"space"})

// Config API

var metricConfigRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "config",
		Name:      "requests_total",
		Help:      "Total of Config API requests.",
	}, []string{"space", "resource", "operation"})
