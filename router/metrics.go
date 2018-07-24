package router

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricEventsReceived)
	prometheus.MustRegister(metricEventsProcessed)
	prometheus.MustRegister(metricEventsDropped)

	prometheus.MustRegister(metricBacklog)
	prometheus.MustRegister(metricProcessingDuration)
}

var metricEventsReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "eventgateway",
		Subsystem: "events",
		Name:      "received_total",
		Help:      "Total of events received.",
	}, []string{"space", "type"})

var metricEventsProcessed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "eventgateway",
		Subsystem: "events",
		Name:      "processed_total",
		Help:      "Total of processed events.",
	}, []string{"space", "type"})

var metricEventsDropped = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "eventgateway",
		Subsystem: "events",
		Name:      "dropped_total",
		Help:      "Total of events dropped due to insufficient processing power.",
	}, []string{"space", "type"})

var metricBacklog = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "eventgateway",
		Subsystem: "events",
		Name:      "backlog",
		Help:      "Gauge of asynchronous events count waiting to be processed.",
	})

var metricProcessingDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "eventgateway",
		Subsystem: "events",
		Name:      "custom_processing_seconds",
		Help: "Bucketed histogram of processing duration of an event. " +
			"From receiving the asynchronous custom event to calling a function.",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 20),
	})

var receivedEventsMutex = sync.Mutex{}
var receivedEvents = map[string]time.Time{}

func reportEventInTheQueue(id string) {
	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	if _, ok := receivedEvents[id]; !ok {
		receivedEvents[id] = time.Now()
	}
}

func reportEventOutOfQueue(id string) {
	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	if startTime, ok := receivedEvents[id]; ok {
		metricProcessingDuration.Observe(time.Since(startTime).Seconds())
		delete(receivedEvents, id)
	}
}
