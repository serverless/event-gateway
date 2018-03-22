package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws/awserr"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/httpapi"
	"github.com/serverless/event-gateway/plugin"
)

// Router calls a target function when an endpoint is hit, and handles pubsub message delivery.
type Router struct {
	sync.Mutex
	targetCache    Targeter
	plugins        *plugin.Manager
	log            *zap.Logger
	workersNumber  uint
	backlogLength  uint
	drain          chan struct{}
	drainWaitGroup sync.WaitGroup
	active         bool
	backlog        chan backlogEvent
}

// New instantiates a new Router
func New(workersNumber uint, backlogLength uint, targetCache Targeter, plugins *plugin.Manager, log *zap.Logger) *Router {
	return &Router{
		targetCache:   targetCache,
		plugins:       plugins,
		log:           log,
		workersNumber: workersNumber,
		backlogLength: backlogLength,
		drain:         make(chan struct{}),
		backlog:       nil,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	// if we're draining requests, spit back a 503
	if router.isDraining() {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "application/json")
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: http.StatusText(http.StatusServiceUnavailable)}}})
		return
	}

	// isHTTPEvent checks if a request carries HTTP event. It also accepts pre-flight CORS requests because CORS is
	// resolved downstream.
	if isHTTPEvent(r) {
		metricEventsReceived.WithLabelValues("", "http").Inc()

		event, _, err := router.eventFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: err.Error()}}})
			return
		}

		router.handleHTTPEvent(event, w, r)
	} else {
		cors.AllowAll().ServeHTTP(w, r, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: "custom event can be emitted only with POST method"}}})
				return
			}

			event, path, err := router.eventFromRequest(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: err.Error()}}})
				return
			}

			if event.Type == eventpkg.TypeInvoke {
				functionID := function.ID(r.Header.Get(headerFunctionID))
				space := r.Header.Get(headerSpace)
				if space == "" {
					space = "default"
				}

				metricEventsReceived.WithLabelValues(space, "invoke").Inc()

				router.handleInvokeEvent(space, functionID, path, event, w)

				metricEventsProcessed.WithLabelValues(space, "invoke").Inc()
			} else if !event.IsSystem() {
				reportReceivedEvent(event.ID)

				router.enqueueWork(path, event)
				w.WriteHeader(http.StatusAccepted)
			}
		})
	}
}

// StartWorkers spins up workerNumber goroutines for processing
// the event subscriptions.
func (router *Router) StartWorkers() {
	router.Lock()
	defer router.Unlock()

	if router.active {
		// the system is already active or being started by another goroutine
		return
	}
	router.active = true

	if router.backlog == nil {
		router.backlog = make(chan backlogEvent, router.backlogLength)
	}

	router.log.Debug("Starting processing workers.", zap.Uint("workers", router.workersNumber), zap.Uint("backlog", router.backlogLength))
	for i := 0; i < int(router.workersNumber); i++ {
		router.drainWaitGroup.Add(1)
		go router.loop()
	}
}

// Drain causes new requests to return 503, and blocks until the work queue is processed.
func (router *Router) Drain() {
	// try to close the draining chan
	router.Lock()
	select {
	case <-router.drain:
		// already closed
	default:
		close(router.drain)
	}
	router.Unlock()

	// wait for children to drain the work queue
	router.drainWaitGroup.Wait()

	router.Lock()
	if router.active {
		router.active = false
	}
	router.Unlock()
}

// WaitForFunction returns a chan that is closed when a function is created.
// Primarily for testing purposes.
func (router *Router) WaitForFunction(space string, id function.ID) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.Function(space, id)
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForEndpoint returns a chan that is closed when an endpoint is created.
// Primarily for testing purposes.
func (router *Router) WaitForEndpoint(method, path string) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			_, res, _, _ := router.targetCache.HTTPBackingFunction(method, path)
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForSubscriber returns a chan that is closed when an event has a subscriber.
// Primarily for testing purposes.
func (router *Router) WaitForSubscriber(path string, eventType eventpkg.Type) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.SubscribersOfEvent(path, eventType)
			if len(res) > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// headerFunctionID is a header name for specifying function id for sync invocation.
const headerFunctionID = "function-id"
const headerSpace = "space"
const hostedDomain = "(eventgateway([a-z-]*)?.io|slsgateway.com)"

var (
	errUnableToLookUpRegisteredFunction = errors.New("unable to look up registered function")
)

func (router *Router) handleHTTPEvent(event *eventpkg.Event, w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	reqMethod := r.Method

	// check if CORS pre-flight request
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		reqMethod = r.Header.Get("Access-Control-Request-Method")
	}
	space, backingFunction, params, corsConfig := router.targetCache.HTTPBackingFunction(
		strings.ToUpper(reqMethod), extractPath(r.Host, r.URL.EscapedPath()),
	)
	if backingFunction == nil {
		router.log.Debug("Function not found for HTTP event.", zap.Object("event", event))
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: "resource not found"}}})
		return
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		httpdata := event.Data.(*eventpkg.HTTPEvent)
		httpdata.Params = params
		event.Data = httpdata
		resp, err := router.callFunction(space, *backingFunction, *event)
		if err != nil {
			message := determineErrorMessage(err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: message}}})
			return
		}

		httpResponse := &HTTPResponse{StatusCode: http.StatusOK}
		err = json.Unmarshal(resp, httpResponse)
		if err != nil {
			router.log.Info("HTTP response object malformed.", zap.String("response", string(resp)))
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: "HTTP response object malformed"}}})
			return
		}

		for key, value := range httpResponse.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(httpResponse.StatusCode)
		resp = []byte(httpResponse.Body)

		_, err = w.Write(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: err.Error()}}})
			return
		}
	}

	if corsConfig == nil {
		handler(w, r)
	} else {
		corsOptions := cors.Options{
			AllowedOrigins:     corsConfig.Origins,
			AllowedHeaders:     corsConfig.Headers,
			AllowedMethods:     corsConfig.Methods,
			AllowCredentials:   corsConfig.AllowCredentials,
			OptionsPassthrough: false,
		}

		cors.New(corsOptions).ServeHTTP(w, r, handler)
	}

	metricEventsProcessed.WithLabelValues(space, "http").Inc()
}

func (router *Router) handleInvokeEvent(space string, functionID function.ID, path string, event *eventpkg.Event, w http.ResponseWriter) {
	encoder := json.NewEncoder(w)

	if !router.targetCache.InvokableFunction(path, space, functionID) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: "function or subscription not found"}}})
		return
	}

	resp, err := router.callFunction(space, functionID, *event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		message := determineErrorMessage(err)
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: message}}})
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: err.Error()}}})
		return
	}
}

func determineErrorMessage(err error) string {
	message := "Function call failed. Please check logs."
	if accessError, ok := err.(*function.ErrFunctionAccessDenied); ok {
		if originalErr, ok := accessError.Original.(awserr.Error); ok {
			switch originalErr.Code() {
			case "AccessDeniedException":
				message = "Function call failed with AccessDeniedException. The provided credentials do not" +
					" have the required IAM permissions to invoke this function. Please attach the" +
					" lambda:invokeFunction permission to these credentials."
			case "UnrecognizedClientException":
				message = "Function call failed with UnrecognizedClientException. The provided credentials" +
					" are invalid. Please provide valid credentials."
			case "ExpiredTokenException":
				message = "Function call failed with ExpiredTokenException. The provided security token for" +
					" the function has expired. Please provide an updated security token or provide" +
					" permanent credentials."
			}
		}
	}

	return message
}

func (router *Router) enqueueWork(path string, event *eventpkg.Event) {
	if event.IsSystem() {
		router.log.Debug("System event received.", zap.Object("event", event))
	}

	select {
	case router.backlog <- backlogEvent{
		path:  path,
		event: *event,
	}:
		metricBacklog.Inc()
	default:
		// We could not submit any work, this is NOT good but we will sacrifice consistency for availability for now.
		metricEventsDropped.WithLabelValues("", "custom").Inc()
	}
}

// callFunction looks up a function and calls it.
func (router *Router) callFunction(space string, backingFunctionID function.ID, event eventpkg.Event) ([]byte, error) {
	router.log.Debug("Invoking function.",
		zap.String("space", space),
		zap.String("functionId", string(backingFunctionID)),
		zap.Object("event", event))
	err := router.emitSystemFunctionInvoking(space, backingFunctionID, event)
	if err != nil {
		router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
			zap.Object("event", event),
			zap.Error(err))
		return nil, err
	}

	// Call the target backing function.
	f := router.targetCache.Function(space, backingFunctionID)
	if f == nil {
		return []byte{}, errUnableToLookUpRegisteredFunction
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	result, err := f.Call(payload)
	if err != nil {
		router.log.Info("Function invocation failed.",
			zap.String("space", space),
			zap.String("functionId", string(backingFunctionID)),
			zap.Object("event", event),
			zap.Error(err))

		router.emitSystemFunctionInvocationFailed(space, backingFunctionID, event, err)
	} else {
		router.log.Debug("Function invoked.",
			zap.String("space", space),
			zap.String("functionId", string(backingFunctionID)),
			zap.Object("event", event),
			zap.ByteString("result", result))

		router.emitSystemFunctionInvoked(space, backingFunctionID, event, payload)
	}

	return result, err
}

// loop is the main loop for a pub/sub worker goroutine
func (router *Router) loop() {
	for {
		// we use three select statements here to give preference
		// to the work chan, but fall-through to exiting when
		// the drain chan is closed and there's nothing to do.

		// 1. see if there's work in a non-blocking way
		select {
		case e := <-router.backlog:
			metricBacklog.Dec()
			router.processEvent(e)
			continue
		default:
		}

		// 2. wait on either work or the drain to close,
		//    blocking on either.
		select {
		case <-router.drain:
			// check AGAIN to make sure there's no work.
			// without this, there is a race condition
			// where we exit before work is processed.
			select {
			case e := <-router.backlog:
				metricBacklog.Dec()
				router.processEvent(e)
				continue
			default:
			}

			// no more work to do, decrement WaitGroup and return
			router.drainWaitGroup.Done()
			return
		case e := <-router.backlog:
			metricBacklog.Dec()
			router.processEvent(e)
		}
	}
}

// processEvent call all functions subscribed for an event
func (router *Router) processEvent(e backlogEvent) {
	reportEventOutOfQueue(e.event.ID)

	subscribers := router.targetCache.SubscribersOfEvent(e.path, e.event.Type)
	for _, subscriber := range subscribers {
		router.callFunction(subscriber.Space, subscriber.ID, e.event)
	}

	metricEventsProcessed.WithLabelValues("", "custom").Inc()
}

func (router *Router) emitSystemEventReceived(path string, event eventpkg.Event, headers map[string]string) error {
	system := eventpkg.New(
		eventpkg.SystemEventReceivedType,
		mimeJSON,
		eventpkg.SystemEventReceivedData{Path: path, Event: event, Headers: headers},
	)
	router.enqueueWork("/", system)
	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoking(space string, functionID function.ID, event eventpkg.Event) error {
	system := eventpkg.New(
		eventpkg.SystemFunctionInvokingType,
		mimeJSON,
		eventpkg.SystemFunctionInvokingData{Space: space, FunctionID: functionID, Event: event},
	)
	router.enqueueWork("/", system)

	metricEventsReceived.WithLabelValues(space, string(eventpkg.SystemFunctionInvokingType)).Inc()

	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoked(space string, functionID function.ID, event eventpkg.Event, result []byte) error {
	system := eventpkg.New(
		eventpkg.SystemFunctionInvokedType,
		mimeJSON,
		eventpkg.SystemFunctionInvokedData{Space: space, FunctionID: functionID, Event: event, Result: result})
	router.enqueueWork("/", system)

	metricEventsReceived.WithLabelValues(space, string(eventpkg.SystemFunctionInvokedType)).Inc()

	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvocationFailed(space string, functionID function.ID, event eventpkg.Event, err error) {
	if _, ok := err.(*function.ErrFunctionError); ok {
		system := eventpkg.New(
			eventpkg.SystemFunctionInvocationFailedType,
			mimeJSON,
			eventpkg.SystemFunctionInvocationFailedData{Space: space, FunctionID: functionID, Event: event, Error: err})
		router.enqueueWork("/", system)

		metricEventsReceived.WithLabelValues(space, string(eventpkg.SystemFunctionInvocationFailedType)).Inc()
	}
}

// isDraining returns true if this Router is being drained of items in its work queue before shutting down.
func (router *Router) isDraining() bool {
	select {
	case <-router.drain:
		return true
	default:
	}
	return false
}

type backlogEvent struct {
	path  string
	event eventpkg.Event
}
