package metrics

import "github.com/prometheus/client_golang/prometheus"

// RequestDuration is a histogram with buckets that
// are incrementally 10% larger than the last, with valid
// values ranging from 0.1 to ~62370
var RequestDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "gateway_request_duration_milliseconds",
		Help:    "Request duration distribution",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.1, 140),
	})

// DroppedPubSubEvents counts the number of times we need
// to drop events instead of forwarding them in the pubsub
// system. This should be alerted on in a monitoring system,
// and trigger adding more capacity.
var DroppedPubSubEvents = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "gateway_pubsub_events_dropped",
		Help: "Dropped events due to insufficient processing power.",
	})
