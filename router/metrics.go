package router

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(routerEventsAsyncReceived)
	prometheus.MustRegister(routerEventsAsyncDropped)
	prometheus.MustRegister(routerEventsAsyncProceeded)
	prometheus.MustRegister(routerEventsSyncReceived)
	prometheus.MustRegister(routerEventsSyncProceeded)
	prometheus.MustRegister(routerBacklog)
	prometheus.MustRegister(routerProcessingDuration)
}

var routerEventsAsyncReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "events_async_received_total",
		Help:      "Total of asynchronously handled events received (including system events).",
	})

var routerEventsAsyncDropped = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "events_async_dropped_total",
		Help:      "Total of asynchronously handled events dropped due to insufficient processing power.",
	})

var routerEventsAsyncProceeded = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "events_async_proceeded_total",
		Help:      "Total of asynchronously proceeded events.",
	})

var routerEventsSyncReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "events_sync_received_total",
		Help:      "Total of synchronously handled (HTTP and invoke) events received.",
	})

var routerEventsSyncProceeded = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "events_sync_proceeded_total",
		Help: "Total of synchronously proceeded events. This counter excludes events for which there was no function " +
			"registered or error occured during processing phase.",
	})

var routerBacklog = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "backlog_events",
		Help:      "Gauge of asynchronous events count waiting to be processed by the router.",
	})

var routerProcessingDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "gateway",
		Subsystem: "router",
		Name:      "event_processing_seconds",
		Help: "Bucketed histogram of processing duration of an event in the router. " +
			"From receiving the asynchronous event to calling a function.",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 20),
	})

var receivedEventsMutex = sync.Mutex{}
var receivedEvents = map[string]time.Time{}

func reportReceivedEvent(id string) {
	routerEventsAsyncReceived.Inc()

	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	receivedEvents[id] = time.Now()
}

func reportProceededEvent(id string) {
	routerEventsAsyncProceeded.Inc()

	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	if startTime, ok := receivedEvents[id]; ok {
		routerProcessingDuration.Observe(time.Since(startTime).Seconds())
		delete(receivedEvents, id)
	}
}
