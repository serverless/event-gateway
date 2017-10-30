package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"go.uber.org/zap"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/plugin"
)

// Router calls a target function when an endpoint is hit, and handles pubsub message delivery.
type Router struct {
	sync.Mutex
	targetCache    Targeter
	plugins        *plugin.Manager
	dropMetric     prometheus.Counter
	log            *zap.Logger
	workerNumber   uint
	drain          chan struct{}
	drainWaitGroup sync.WaitGroup
	active         bool
	work           chan workEvent
}

// New instantiates a new Router
func New(targetCache Targeter, plugins *plugin.Manager, dropMetric prometheus.Counter, log *zap.Logger) *Router {
	return &Router{
		targetCache:  targetCache,
		plugins:      plugins,
		dropMetric:   dropMetric,
		log:          log,
		workerNumber: 20,
		drain:        make(chan struct{}),
		work:         nil,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we're draining requests, spit back a 503
	if router.isDraining() {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	// isHTTPEvent checks if a request carries HTTP event. It also accepts pre-flight CORS requests because CORS is
	// resolved downstream.
	if isHTTPEvent(r) {
		event, _, err := router.eventFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		router.handleHTTPEvent(event, w, r)
	} else {
		cors.AllowAll().ServeHTTP(w, r, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "custom event can be emitted only with POST method")
				return
			}

			event, path, err := router.eventFromRequest(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if event.Type == eventpkg.TypeInvoke {
				router.handleInvokeEvent(event, w, r)
			} else if !event.IsSystem() {
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

	if router.work == nil {
		router.work = make(chan workEvent, router.workerNumber*2)
	}

	for i := 0; i < int(router.workerNumber); i++ {
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
func (router *Router) WaitForFunction(id functions.FunctionID) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.Function(id)
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
			res, _, _ := router.targetCache.HTTPBackingFunction(method, path)
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
const hostedDomain = "eventgateway([a-z-]*)?.io"

var (
	errUnableToLookUpRegisteredFunction = errors.New("unable to look up registered function")
)

func (router *Router) eventFromRequest(r *http.Request) (*eventpkg.Event, string, error) {
	path := extractPath(r.Host, r.URL.Path)
	eventType := extractEventType(r)

	mime := r.Header.Get("content-type")
	if mime == "" {
		mime = mimeOctetStrem
	}

	body := []byte{}
	var err error
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, "", err
		}
	}

	event := eventpkg.NewEvent(eventType, mime, body)
	if mime == mimeJSON && len(body) > 0 {
		err = json.Unmarshal(body, &event.Data)
		if err != nil {
			return nil, "", errors.New("malformed JSON body")
		}
	}

	if event.Type == eventpkg.TypeHTTP {
		event.Data = &eventpkg.HTTPEvent{
			Headers: r.Header,
			Query:   r.URL.Query(),
			Body:    event.Data,
			Host:    r.Host,
			Path:    r.URL.Path, // it's not path var as user has to get path from request to platform's EG
			Method:  r.Method,
		}
	}

	router.log.Debug("Event received.", zap.String("path", path), zap.Object("event", event))
	err = router.emitSystemEventReceived(path, *event, r.Header)
	if err != nil {
		router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
			zap.Object("event", event),
			zap.Error(err))
		return nil, "", err
	}

	return event, path, nil
}
func (router *Router) handleHTTPEvent(event *eventpkg.Event, w http.ResponseWriter, r *http.Request) {
	reqMethod := r.Method

	// check if CORS pre-flight request
	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		reqMethod = r.Header.Get("Access-Control-Request-Method")
	}

	backingFunction, params, corsConfig := router.targetCache.HTTPBackingFunction(
		strings.ToUpper(reqMethod), extractPath(r.Host, r.URL.EscapedPath()),
	)
	if backingFunction == nil {
		router.log.Debug("Function not found for HTTP event.", zap.Object("event", event))
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		httpdata := event.Data.(*eventpkg.HTTPEvent)
		httpdata.Params = params
		event.Data = httpdata
		resp, err := router.callFunction(*backingFunction, *event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		httpResponse := &HTTPResponse{StatusCode: http.StatusOK}
		err = json.Unmarshal(resp, httpResponse)
		if err != nil {
			httperr := NewErrHTTPResponseObjectMalformed()
			http.Error(w, httperr.Error(), httperr.StatusCode)

			router.log.Info(httperr.Error(), zap.String("response", string(resp)))

			return
		}

		for key, value := range httpResponse.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(httpResponse.StatusCode)
		resp = []byte(httpResponse.Body)

		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
}

func (router *Router) handleInvokeEvent(event *eventpkg.Event, w http.ResponseWriter, r *http.Request) {
	functionID := functions.FunctionID(r.Header.Get(headerFunctionID))
	resp, err := router.callFunction(functionID, *event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (router *Router) enqueueWork(path string, event *eventpkg.Event) {
	if event.IsSystem() {
		router.log.Debug("System event received.", zap.Object("event", event))
	}

	select {
	case router.work <- workEvent{
		path:  path,
		event: *event,
	}:
	default:
		// We could not submit any work, this is NOT good but we will sacrifice consistency for availability for now.
		router.dropMetric.Inc()
	}
}

// callFunction looks up a function and calls it.
func (router *Router) callFunction(backingFunctionID functions.FunctionID, event eventpkg.Event) ([]byte, error) {
	backingFunction := router.targetCache.Function(backingFunctionID)
	if backingFunction == nil {
		return []byte{}, errUnableToLookUpRegisteredFunction
	}

	var chosenFunction = backingFunction.ID
	if backingFunction.Provider.Type == functions.Weighted {
		chosen, err := backingFunction.Provider.Weighted.Choose()
		if err != nil {
			return nil, err
		}
		chosenFunction = chosen
	}

	router.log.Debug("Invoking function.", zap.String("functionId", string(backingFunctionID)), zap.Object("event", event))
	err := router.emitSystemFunctionInvoking(backingFunctionID, event)
	if err != nil {
		router.log.Debug("Event processing stopped because sync plugin subscription returned an error.",
			zap.Object("event", event),
			zap.Error(err))
		return nil, err
	}

	// Call the target backing function.
	f := router.targetCache.Function(chosenFunction)
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
			zap.String("functionId", string(backingFunctionID)), zap.Object("event", event), zap.Error(err))

		router.emitSystemFunctionInvocationFailed(backingFunctionID, event, err)
	} else {
		router.log.Debug("Function invoked.",
			zap.String("functionId", string(backingFunctionID)), zap.Object("event", event), zap.ByteString("result", result))

		router.emitSystemFunctionInvoked(backingFunctionID, event, payload)
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
		case e := <-router.work:
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
			case e := <-router.work:
				router.processEvent(e)
				continue
			default:
			}

			// no more work to do, decrement WaitGroup and return
			router.drainWaitGroup.Done()
			return
		case e := <-router.work:
			router.processEvent(e)
		}
	}
}

// processEvent call all functions subscribed for an event
func (router *Router) processEvent(e workEvent) {
	subscribers := router.targetCache.SubscribersOfEvent(e.path, e.event.Type)
	for _, subscriber := range subscribers {
		router.callFunction(subscriber, e.event)
	}
}

func (router *Router) emitSystemEventReceived(path string, event eventpkg.Event, headers http.Header) error {
	system := eventpkg.NewEvent(
		eventpkg.SystemEventReceivedType,
		mimeJSON,
		eventpkg.SystemEventReceivedData{Path: path, Event: event, Headers: headers},
	)
	router.enqueueWork("/", system)
	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoking(functionID functions.FunctionID, event eventpkg.Event) error {
	system := eventpkg.NewEvent(
		eventpkg.SystemFunctionInvokingType,
		mimeJSON,
		eventpkg.SystemFunctionInvokingData{FunctionID: functionID, Event: event},
	)
	router.enqueueWork("/", system)
	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvoked(functionID functions.FunctionID, event eventpkg.Event, result []byte) error {
	system := eventpkg.NewEvent(
		eventpkg.SystemFunctionInvokedType,
		mimeJSON,
		eventpkg.SystemFunctionInvokedData{FunctionID: functionID, Event: event, Result: result})
	router.enqueueWork("/", system)
	return router.plugins.React(system)
}

func (router *Router) emitSystemFunctionInvocationFailed(functionID functions.FunctionID, event eventpkg.Event, err error) {
	if _, ok := err.(*functions.ErrFunctionError); ok {
		system := eventpkg.NewEvent("gateway.function.invocationFailed", mimeJSON, struct {
			FunctionID string `json:"functionId"`
		}{string(functionID)})

		router.enqueueWork("/", system)
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

func extractPath(host, path string) string {
	extracted := path
	rxp, _ := regexp.Compile(hostedDomain)
	if rxp.MatchString(host) {
		extracted = "/" + strings.Split(host, ".")[0] + extracted
	}
	return extracted
}

func extractEventType(r *http.Request) eventpkg.Type {
	eventType := eventpkg.Type(r.Header.Get("event"))
	if eventType == "" {
		eventType = eventpkg.TypeHTTP
	}
	return eventType
}
