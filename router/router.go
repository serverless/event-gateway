package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	eventpkg "github.com/serverless/event-gateway/event"
	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/internal/cache"
)

// Router calls a target function when an endpoint is hit, and handles pubsub message delivery.
type Router struct {
	sync.Mutex
	targetCache          cache.Targeter
	dropMetric           prometheus.Counter
	log                  *zap.Logger
	NWorkers             uint
	drain                chan struct{}
	drainWaitGroup       sync.WaitGroup
	active               bool
	work                 chan event
	responseWriteTimeout time.Duration
}

// New instantiates a new Router
func New(targetCache cache.Targeter, dropMetric prometheus.Counter, log *zap.Logger) *Router {
	return &Router{
		targetCache: targetCache,
		dropMetric:  dropMetric,
		log:         log,
		NWorkers:    20,
		drain:       make(chan struct{}),
		work:        nil,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we're draining requests, spit back a 503
	if router.isDraining() {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	event, err := fromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if event.Type == eventpkg.TypeHTTP || event.Type == eventpkg.TypeInvoke {
		router.handleSyncEvent(event, payload, w, r)
	} else if r.Method == http.MethodPost {
		router.enqueueWork(r.URL.Path, event.Type, payload)
		w.WriteHeader(http.StatusAccepted)
	}
}

// StartWorkers spins up NWorkers goroutines for processing
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
		router.work = make(chan event, router.NWorkers*2)
	}

	for i := 0; i < int(router.NWorkers); i++ {
		router.drainWaitGroup.Add(1)
		go router.loop()
	}
}

// Drain causes new requests to return 503, and blocks until
// the work queue is processed.
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
			res, _ := router.targetCache.HTTPBackingFunction(method, path)
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

const (
	// headerFunctionID is a header name for specifying function id for sync invocation.
	headerFunctionID = "function-id"

	internalFunctionError = "gateway.info.functionError"
)

var (
	errUnableToLookUpRegisteredFunction = errors.New("unable to look up registered function")
)

func (router *Router) handleSyncEvent(event *eventpkg.Event, payload []byte, w http.ResponseWriter, r *http.Request) {
	router.log.Debug("Event received.", zap.String("event", string(payload)))

	var resp []byte
	var functionID functions.FunctionID

	if event.Type == eventpkg.TypeInvoke {
		functionID = functions.FunctionID(r.Header.Get(headerFunctionID))
	} else if event.Type == eventpkg.TypeHTTP {
		backingFunction, params := router.targetCache.HTTPBackingFunction(strings.ToUpper(r.Method), r.URL.EscapedPath())
		if backingFunction == nil {
			router.log.Debug("Function not found for HTTP event.", zap.String("event", string(payload)))
			http.Error(w, "Resource not found", http.StatusNotFound)
			return
		}

		httpdata := event.Data.(*eventpkg.HTTPEvent)
		httpdata.Params = params
		event.Data = httpdata
		var err error
		payload, err = json.Marshal(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		functionID = *backingFunction
	}

	router.log.Debug("Function triggered.", zap.String("functionId", string(functionID)), zap.String("event", string(payload)))

	resp, err := router.callFunction(functionID, payload)
	if err != nil {
		router.log.Info("Function invocation failed.",
			zap.String("functionId", string(functionID)), zap.String("event", string(payload)), zap.Error(err))

		http.Error(w, err.Error(), http.StatusInternalServerError)
		router.emitFunctionErrorEvent(functionID, payload, err)
		return
	}

	router.log.Debug("Function finished.",
		zap.String("functionId", string(functionID)), zap.String("event", string(payload)),
		zap.String("response", string(resp)))

	if event.Type == eventpkg.TypeHTTP {
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
	}

	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (router *Router) enqueueWork(path string, eventType eventpkg.Type, payload []byte) {
	router.log.Debug("Event received.", zap.String("path", path), zap.String("event", string(payload)))

	select {
	case router.work <- event{
		path:      path,
		eventType: eventType,
		payload:   payload,
	}:
	default:
		// We could not submit any work, this is NOT good but we will sacrifice consistency for availability for now.
		router.dropMetric.Inc()
	}
}

// callFunction looks up a function and calls it.
func (router *Router) callFunction(backingFunctionID functions.FunctionID, payload []byte) ([]byte, error) {
	backingFunction := router.targetCache.Function(backingFunctionID)
	if backingFunction == nil {
		return []byte{}, errUnableToLookUpRegisteredFunction
	}

	var chosenFunction = backingFunction.ID
	if backingFunction.Provider.Type == functions.Weighted {
		chosen, err := backingFunction.Provider.Weighted.Choose()
		if err != nil {
			return []byte{}, err
		}
		chosenFunction = chosen
	}

	// Call the target backing function.
	f := router.targetCache.Function(chosenFunction)
	if f == nil {
		return []byte{}, errUnableToLookUpRegisteredFunction
	}

	return f.Call(payload)
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
func (router *Router) processEvent(e event) {
	subscribers := router.targetCache.SubscribersOfEvent(e.path, e.eventType)
	for _, subscriber := range subscribers {
		router.log.Debug("Function triggered.",
			zap.String("functionId", string(subscriber)), zap.String("path", e.path), zap.String("event", string(e.payload)))

		resp, err := router.callFunction(subscriber, e.payload)

		if err != nil {
			router.log.Info("Function invocation failed.",
				zap.String("functionId", string(subscriber)), zap.String("path", e.path), zap.String("event", string(e.payload)), zap.Error(err))

			router.emitFunctionErrorEvent(subscriber, e.payload, err)
		} else {
			router.log.Debug("Function finished.",
				zap.String("functionId", string(subscriber)), zap.String("path", e.path), zap.String("event", string(e.payload)),
				zap.String("response", string(resp)))
		}
	}
}

func (router *Router) emitFunctionErrorEvent(functionID functions.FunctionID, payload []byte, err error) {
	if _, ok := err.(*functions.ErrFunctionError); ok {
		internal := eventpkg.NewEvent(internalFunctionError, mimeJSON, struct {
			FunctionID string `json:"functionId"`
		}{string(functionID)})
		payload, err = json.Marshal(internal)
		if err == nil {
			router.enqueueWork("/", internal.Type, payload)
		}
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
