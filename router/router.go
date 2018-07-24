package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws/awserr"
	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/function"
	"github.com/serverless/event-gateway/httpapi"
	ihttp "github.com/serverless/event-gateway/internal/http"
	"github.com/serverless/event-gateway/plugin"
)

const (
	mimeJSON = "application/json"
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

	reqMethod := r.Method
	// check if CORS pre-flight request
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		reqMethod = r.Header.Get("Access-Control-Request-Method")
	}
	path := extractPath(r.Host, r.URL.EscapedPath())

	handler := func(w http.ResponseWriter, r *http.Request) {
		event, err := eventpkg.FromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			encoder.Encode(&httpapi.Response{Errors: []httpapi.Error{{Message: err.Error()}}})
			return
		}
		if event.IsSystem() { // System event can only be emitted from inside EG
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		router.log.Debug("Event received.", zap.String("path", path), zap.Object("event", event))
		err = router.emitSystemEventReceived(path, *event, r.Header)
		if err != nil {
			router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
				zap.Object("event", event),
				zap.Error(err))
			return
		}

		syncSubscriber := router.targetCache.SyncSubscriber(r.Method, path, event.EventType)
		if syncSubscriber != nil { // There is sync subscriber and possibly async subscribers also
			router.handleSyncSubscription(path, *event, *syncSubscriber, w, r)
		}

		router.handleAsyncSubscriptions(r.Method, path, *event, r)
		if syncSubscriber == nil {
			w.WriteHeader(http.StatusAccepted)
		}
	}

	corsConfig := router.targetCache.CORS(reqMethod, path)
	if corsConfig != nil {
		corsOptions := cors.Options{
			AllowedOrigins:     corsConfig.AllowedOrigins,
			AllowedHeaders:     corsConfig.AllowedHeaders,
			AllowedMethods:     corsConfig.AllowedMethods,
			AllowCredentials:   corsConfig.AllowCredentials,
			OptionsPassthrough: false,
		}

		cors.New(corsOptions).ServeHTTP(w, r, handler)
	} else {
		handler(w, r)
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

var (
	errUnableToLookUpRegisteredFunction = errors.New("unable to look up registered function")
)

func (router *Router) handleSyncSubscription(path string, event eventpkg.Event, subscriber SyncSubscriber, w http.ResponseWriter, r *http.Request) {
	// metrics & logs
	metricEventsReceived.WithLabelValues(subscriber.Space, string(event.EventType)).Inc()
	router.log.Debug("Event received.", zap.String("path", path), zap.String("space", subscriber.Space), zap.Object("event", event))
	err := router.emitSystemEventReceived(path, event, r.Header)
	if err != nil {
		router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
			zap.Object("event", event),
			zap.Error(err))
		return
	}

	err = router.authorizeEventType(subscriber.Space, &event, r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// add params to HTTP Request object
	if event.EventType == eventpkg.TypeHTTPRequest {
		httpRequestData := event.Data.(*eventpkg.HTTPRequestData)
		httpRequestData.Params = subscriber.Params
		event.Data = httpRequestData
	}
	router.httpRequestHandler(subscriber.Space, subscriber.FunctionID, &event)(w, r)

	metricEventsProcessed.WithLabelValues(subscriber.Space, string(event.EventType)).Inc()
}

// Return http.HandlerFunc that will call remote function and return response in HTTP response object.
func (router *Router) httpRequestHandler(space string, backingFunction function.ID, event *eventpkg.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encoder := json.NewEncoder(w)

		resp, err := router.callFunction(space, backingFunction, *event)
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
}

// handleAsyncSubscriptions fetched events subscribers, runs authorization and enqueues event in the queue
func (router *Router) handleAsyncSubscriptions(method, path string, event eventpkg.Event, r *http.Request) {
	if event.IsSystem() {
		router.log.Debug("System event received.", zap.Object("event", event))
	}

	subscribers := router.targetCache.AsyncSubscribers(method, path, event.EventType)
	for _, subscriber := range subscribers {
		metricEventsReceived.WithLabelValues(subscriber.Space, "custom").Inc()

		subEvent := eventpkg.Event{}
		copier.Copy(&subEvent, &event)
		err := router.authorizeEventType(subscriber.Space, &subEvent, r)
		if err == nil {
			router.enqueueWork(method, path, subscriber.Space, subscriber.FunctionID, subEvent)
		}
	}
}

func (router *Router) enqueueWork(method, path, space string, functionID function.ID, event eventpkg.Event) {
	reportEventInTheQueue(event.EventID)

	select {
	case router.backlog <- backlogEvent{
		method:     method,
		path:       path,
		space:      space,
		functionID: functionID,
		event:      event,
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

func (router *Router) authorizeEventType(space string, event *eventpkg.Event, r *http.Request) error {
	eventType := router.targetCache.EventType(space, event.EventType)
	if eventType != nil && eventType.AuthorizerID != nil {
		payload := AuthorizerPayload{
			Request: *eventpkg.NewHTTPRequestData(r, nil),
			Event:   *event,
		}
		resp, err := router.callAuthorizer(space, *eventType.AuthorizerID, payload)
		if err != nil {
			return err
		}

		authorizerResponse := &AuthorizerResponse{}
		err = json.Unmarshal(resp, authorizerResponse)
		if err != nil {
			router.log.Info("Failed to unmarshal authorizer function response.",
				zap.ByteString("response", resp),
				zap.String("space", space),
				zap.String("authorizerId", string(*eventType.AuthorizerID)),
				zap.Object("event", event))
			return err
		}

		if authorizerResponse.AuthorizationError != nil {
			router.log.Info("Authorization failed.",
				zap.String("error", authorizerResponse.AuthorizationError.Message),
				zap.String("space", space),
				zap.String("authorizerId", string(*eventType.AuthorizerID)),
				zap.Object("event", event))
			return errors.New(authorizerResponse.AuthorizationError.Message)
		}

		if egExternsions, ok := event.Extensions["eventgateway"]; ok {
			egExternsions.(map[string]interface{})["authorization"] = authorizerResponse.Authorization
		} else {
			event.Extensions = map[string]interface{}{
				"eventgateway": map[string]interface{}{
					"authorization": authorizerResponse.Authorization,
				},
			}
		}
	}

	return nil
}

// callAuthorizer looks up an authorizer function and calls it.
func (router *Router) callAuthorizer(space string, backingFunctionID function.ID, payload AuthorizerPayload) ([]byte, error) {
	router.log.Debug("Invoking authorizer function.",
		zap.String("space", space),
		zap.String("functionId", string(backingFunctionID)))

	// Call the target backing function.
	f := router.targetCache.Function(space, backingFunctionID)
	if f == nil {
		return []byte{}, errUnableToLookUpRegisteredFunction
	}

	callPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	result, err := f.Call(callPayload)
	if err != nil {
		router.log.Info("Authorizer function invocation failed.",
			zap.String("space", space),
			zap.String("functionId", string(backingFunctionID)),
			zap.Error(err))
	} else {
		router.log.Debug("Authorizer function invoked.",
			zap.String("space", space),
			zap.String("functionId", string(backingFunctionID)),
			zap.ByteString("result", result))
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
	reportEventOutOfQueue(e.event.EventID)

	router.callFunction(e.space, e.functionID, e.event)

	metricEventsProcessed.WithLabelValues(e.space, "custom").Inc()
}

func (router *Router) emitSystemEventReceived(path string, event eventpkg.Event, header http.Header) error {
	system := eventpkg.New(
		eventpkg.SystemEventReceivedType,
		mimeJSON,
		eventpkg.SystemEventReceivedData{Path: path, Event: event, Headers: ihttp.FlattenHeader(header)},
	)
	router.handleAsyncSubscriptions(http.MethodPost, "/", *system, nil)
	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoking(space string, functionID function.ID, event eventpkg.Event) error {
	system := eventpkg.New(
		eventpkg.SystemFunctionInvokingType,
		mimeJSON,
		eventpkg.SystemFunctionInvokingData{Space: space, FunctionID: functionID, Event: event},
	)
	router.handleAsyncSubscriptions(http.MethodPost, systemEventPath(space), *system, nil)

	metricEventsReceived.WithLabelValues(space, string(eventpkg.SystemFunctionInvokingType)).Inc()

	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoked(space string, functionID function.ID, event eventpkg.Event, result []byte) error {
	system := eventpkg.New(
		eventpkg.SystemFunctionInvokedType,
		mimeJSON,
		eventpkg.SystemFunctionInvokedData{Space: space, FunctionID: functionID, Event: event, Result: result})
	router.handleAsyncSubscriptions(http.MethodPost, systemEventPath(space), *system, nil)

	metricEventsReceived.WithLabelValues(space, string(eventpkg.SystemFunctionInvokedType)).Inc()

	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvocationFailed(space string, functionID function.ID, event eventpkg.Event, err error) {
	if _, ok := err.(*function.ErrFunctionError); ok {
		system := eventpkg.New(
			eventpkg.SystemFunctionInvocationFailedType,
			mimeJSON,
			eventpkg.SystemFunctionInvocationFailedData{Space: space, FunctionID: functionID, Event: event, Error: err})
		router.handleAsyncSubscriptions(http.MethodPost, systemEventPath(space), *system, nil)

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

type backlogEvent struct {
	space      string
	functionID function.ID
	method     string
	path       string
	event      eventpkg.Event
}
