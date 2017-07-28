package router

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/functions"
	"github.com/serverless/event-gateway/pubsub"
	"github.com/serverless/event-gateway/targetcache"
)

// Router calls a target function when an endpoint is hit, and
// handles pubsub message delivery.
type Router struct {
	sync.Mutex
	targetCache          targetcache.TargetCache
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
func New(targetCache targetcache.TargetCache, dropMetric prometheus.Counter, log *zap.Logger) *Router {
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
	}

	reqBuf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	eventHeader := r.Header.Get("event")
	if eventHeader == "" {
		endpointID := pubsub.NewEndpointID(strings.ToUpper(r.Method), r.URL.EscapedPath())
		router.log.Debug("router serving request", zap.String("endpoint", string(endpointID)))

		res, err := router.callEndpoint(endpointID, reqBuf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == http.MethodPost && r.URL.Path == "/" {
		router.processEvent(event{
			topics:  []pubsub.TopicID{pubsub.TopicID(eventHeader)},
			payload: reqBuf,
		})

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

// WaitForEndpoint returns a chan that is closed when an endpoint is created.
// Primarily for testing purposes.
func (router *Router) WaitForEndpoint(endpointID pubsub.EndpointID) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.BackingFunction(endpointID)
			if res != nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// WaitForSubscriber returns a chan that is closed when a topic has a subscriber.
// Primarily for testing purposes.
func (router *Router) WaitForSubscriber(topic pubsub.TopicID) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := router.targetCache.SubscribersOfTopic(topic)
			if len(res) > 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		close(updatedChan)
	}()
	return updatedChan
}

// callEndpoint determines which function to call when an endpoint is hit.
func (router *Router) callEndpoint(endpointID pubsub.EndpointID, payload []byte) ([]byte, error) {
	// Figure out what function we're targeting.
	backingFunction := router.targetCache.BackingFunction(endpointID)
	if backingFunction == nil {
		retErr := errors.New("for endpoint ID:" + string(endpointID) + ", could not find backing function")
		return []byte{}, retErr
	}

	return router.callFunction(*backingFunction, payload)
}

// callFunction looks up a function and calls it.
func (router *Router) callFunction(backingFunctionID functions.FunctionID, payload []byte) ([]byte, error) {
	backingFunction := router.targetCache.Function(backingFunctionID)
	if backingFunction == nil {
		resErr := errors.New("unable to look up backing function: " + string(backingFunctionID))
		return []byte{}, resErr
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
		resErr := errors.New("unable to look up backing function: " + string(chosenFunction))
		return []byte{}, resErr
	}

	result, err := f.Call(payload)
	if err != nil {
		resErr := errors.New("unable to reach backing function: " + err.Error())
		return []byte{}, resErr
	}

	return result, nil
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

// processEvent sends event to a set of topics,
// and for each of the functions that get called
// as part of those topics.
func (router *Router) processEvent(e event) {
	for _, topicID := range e.topics {
		subscribers := router.targetCache.SubscribersOfTopic(topicID)
		for _, subscriber := range subscribers {
			router.callFunction(subscriber, e.payload)
		}
	}
}

func (router *Router) enqueueWork(topicMap map[pubsub.TopicID]struct{}, payload []byte) {
	if len(topicMap) == 0 {
		return
	}
	topics := []pubsub.TopicID{}
	for topic := range topicMap {
		topics = append(topics, topic)
	}
	select {
	case router.work <- event{
		topics:  topics,
		payload: payload,
	}:
	default:
		// We could not submit any work, this is NOT good but
		// we will sacrifice consistency for availability for now.
		router.dropMetric.Inc()
	}
}

// isDraining returns true if this Router is being drained of items
// in its work queue before shutting down.
func (router *Router) isDraining() bool {
	select {
	case <-router.drain:
		return true
	default:
	}
	return false
}
