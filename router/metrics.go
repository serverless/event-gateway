package router

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(metricSystemFunctionInvokingReceived)
	prometheus.MustRegister(metricSystemFunctionInvokedReceived)
	prometheus.MustRegister(metricSystemFunctionInvocationFailedReceived)

	prometheus.MustRegister(metricEventsAsyncReceived)
	prometheus.MustRegister(metricEventsAsyncDropped)
	prometheus.MustRegister(metricEventsAsyncProcessed)
	prometheus.MustRegister(metricEventsInvokeReceived)
	prometheus.MustRegister(metricEventsInvokeProcessed)
	prometheus.MustRegister(metricBacklog)
	prometheus.MustRegister(metricProcessingDuration)
}

var metricSystemFunctionInvokingReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "system_function_invoking_received_total",
		Help:      "Total of gateway.function.invoking events received.",
	}, []string{"space"})

var metricSystemFunctionInvokedReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "system_function_invoked_received_total",
		Help:      "Total of gateway.function.invoked events received.",
	}, []string{"space"})

var metricSystemFunctionInvocationFailedReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "system_function_invocation_failed_received_total",
		Help:      "Total of gateway.function.invocationFailed events received.",
	}, []string{"space"})

var metricEventsAsyncReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "async_received_total",
		Help:      "Total of asynchronously handled events received (including system events).",
	})

var metricEventsAsyncDropped = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "async_dropped_total",
		Help:      "Total of asynchronously handled events dropped due to insufficient processing power.",
	})

var metricEventsAsyncProcessed = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "async_processed_total",
		Help:      "Total of asynchronously processed events.",
	})

var metricEventsInvokeReceived = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "invoke_received_total",
		Help:      "Total of Invoke events received.",
	}, []string{"space"})

var metricEventsInvokeProcessed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "invoke_processed_total",
		Help: "Total of Invoke events processed. This counter excludes events for which there was no function " +
			"registered or error occured during processing phase.",
	}, []string{"space"})

var metricEventsHTTPReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "http_received_total",
		Help:      "Total of HTTP events received.",
	})

var metricEventsHTTPProcessed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "http_processed_total",
		Help: "Total of HTTP events processed. This counter excludes events for which there was no function " +
			"registered or error occured during processing phase.",
	}, []string{"space"})

var metricBacklog = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "backlog",
		Help:      "Gauge of asynchronous events count waiting to be processed.",
	})

var metricProcessingDuration = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "async_processing_seconds",
		Help: "Bucketed histogram of processing duration of an event. " +
			"From receiving the asynchronous event to calling a function.",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 20),
	})

var receivedEventsMutex = sync.Mutex{}
var receivedEvents = map[string]time.Time{}

func reportReceivedEvent(id string) {
	metricEventsAsyncReceived.Inc()

	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	receivedEvents[id] = time.Now()
}

func reportEventOutOfQueue(id string) {
	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	if startTime, ok := receivedEvents[id]; ok {
		metricProcessingDuration.Observe(time.Since(startTime).Seconds())
		delete(receivedEvents, id)
	}
}
