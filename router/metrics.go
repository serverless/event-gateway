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
	prometheus.MustRegister(metricEventsAsyncProceeded)
	prometheus.MustRegister(metricEventsInvokeReceived)
	prometheus.MustRegister(metricEventsInvokeProceeded)
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

var metricEventsAsyncProceeded = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "async_proceeded_total",
		Help:      "Total of asynchronously proceeded events.",
	})

var metricEventsInvokeReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "invoke_received_total",
		Help:      "Total of Invoke events received.",
	})

var metricEventsInvokeProceeded = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "invoke_proceeded_total",
		Help: "Total of Invoke events proceeded. This counter excludes events for which there was no function " +
			"registered or error occured during processing phase.",
	})

var metricEventsHTTPReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "http_received_total",
		Help:      "Total of HTTP events received.",
	})

var metricEventsHTTPProceeded = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "gateway",
		Subsystem: "events",
		Name:      "http_proceeded_total",
		Help: "Total of HTTP events proceeded. This counter excludes events for which there was no function " +
			"registered or error occured during processing phase.",
	})

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

func reportProceededEvent(id string) {
	metricEventsAsyncProceeded.Inc()

	receivedEventsMutex.Lock()
	defer receivedEventsMutex.Unlock()
	if startTime, ok := receivedEvents[id]; ok {
		metricProcessingDuration.Observe(time.Since(startTime).Seconds())
		delete(receivedEvents, id)
	}
}
