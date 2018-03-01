package httpapi

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricFunctionRegistered)
	prometheus.MustRegister(metricFunctionDeleted)

	prometheus.MustRegister(metricFunctionGetRequests)
	prometheus.MustRegister(metricFunctionRegisterRequests)
	prometheus.MustRegister(metricFunctionDeleteRequests)
	prometheus.MustRegister(metricFunctionUpdateRequests)
	prometheus.MustRegister(metricFunctionListRequests)

	prometheus.MustRegister(metricSubscriptionsCreated)
	prometheus.MustRegister(metricSubscriptionsDeleted)

	prometheus.MustRegister(metricSubscriptionGetRequests)
	prometheus.MustRegister(metricSubscriptionCreateRequests)
	prometheus.MustRegister(metricSubscriptionDeleteRequests)
	prometheus.MustRegister(metricSubscriptionListRequests)
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

// Functions Config API

var metricFunctionGetRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "function_get_requests_total",
		Help:      "Total of Config API get function requests.",
	}, []string{"space"})

var metricFunctionRegisterRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "function_register_requests_total",
		Help:      "Total of Config API register function requests.",
	}, []string{"space"})

var metricFunctionDeleteRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "function_delete_requests_total",
		Help:      "Total of Config API delete function requests.",
	}, []string{"space"})

var metricFunctionUpdateRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "function_update_requests_total",
		Help:      "Total of Config API update function requests.",
	}, []string{"space"})

var metricFunctionListRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "function_list_requests_total",
		Help:      "Total of Config API list functions requests.",
	}, []string{"space"})

// Subscriptions

var metricSubscriptionsCreated = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "subscriptions",
		Name:      "created_total",
		Help:      "Total of subscriptions created.",
	}, []string{"space"})

var metricSubscriptionsDeleted = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "subscriptions",
		Name:      "deleted_total",
		Help:      "Total of subscriptions deleted.",
	}, []string{"space"})

// Subscriptions Config API

var metricSubscriptionGetRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "subscription_get_requests_total",
		Help:      "Total of Config API get subscription requests.",
	}, []string{"space"})

var metricSubscriptionCreateRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "subscription_create_requests_total",
		Help:      "Total of Config API create subscription requests.",
	}, []string{"space"})

var metricSubscriptionDeleteRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "subscription_delete_requests_total",
		Help:      "Total of Config API delete subscription requests.",
	}, []string{"space"})

var metricSubscriptionListRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "httpapi",
		Name:      "subscription_list_requests_total",
		Help:      "Total of Config API list subscriptions requests.",
	}, []string{"space"})
