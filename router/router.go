package router

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/serverless/event-gateway/endpoints"
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
	work                 chan work
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

// IsDraining returns true if this Router is being drained of items
// in its work queue before shutting down.
func (router *Router) IsDraining() bool {
	select {
	case <-router.drain:
		return true
	default:
	}
	return false
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we're draining requests, spit back a 503
	if router.IsDraining() {
		http.Error(w, http.StatusText(503), 503)
	}

	rawPath := r.URL.EscapedPath()
	isAsync := strings.HasSuffix(rawPath, "/async")
	trimmedSuffixPath := strings.TrimSuffix(rawPath, "/async")
	trimmedPath := strings.TrimPrefix(trimmedSuffixPath, "/")
	id := strings.ToUpper(r.Method) + "-" + trimmedPath
	endpointID := endpoints.EndpointID(id)
	router.log.Debug("router serving request", zap.String("endpoint", string(endpointID)))

	reqBuf, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Fprintf(w, "%s", err)
	}

	if isAsync {
		w.WriteHeader(http.StatusAccepted)
		// TODO use goroutine pool to avoid unbounded goroutine creation
		go router.CallEndpoint(endpointID, reqBuf)
	} else {
		res, err := router.CallEndpoint(endpointID, reqBuf)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = w.Write(res)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func (router *Router) subscribers(fnGroup *functions.FunctionID, chosenFunction functions.FunctionID) (
	map[pubsub.TopicID]struct{}, map[pubsub.TopicID]struct{}) {

	sendInput := map[pubsub.TopicID]struct{}{}
	sendOutput := map[pubsub.TopicID]struct{}{}

	fillInputMap := func(fid functions.FunctionID) {
		inputDest := router.targetCache.FunctionInputToTopics(fid)

		for _, topic := range inputDest {
			sendInput[topic] = struct{}{}
		}
	}

	fillOutputMap := func(fid functions.FunctionID) {
		outputDest := router.targetCache.FunctionOutputToTopics(fid)

		for _, topic := range outputDest {
			sendOutput[topic] = struct{}{}
		}
	}

	if fnGroup != nil {
		fillInputMap(*fnGroup)
		fillOutputMap(*fnGroup)
	}

	fillInputMap(chosenFunction)
	fillOutputMap(chosenFunction)

	return sendInput, sendOutput
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
	case router.work <- work{
		topics:  topics,
		payload: payload,
	}:
	default:
		// We could not submit any work, this is NOT good but
		// we will sacrifice consistency for availability for now.
		router.dropMetric.Inc()
	}
}

// CallEndpoint determines which function to call when an endpoint is hit, and
// submits pubsub events to the work queue.
func (router *Router) CallEndpoint(endpointID endpoints.EndpointID, payload []byte) ([]byte, error) {
	// Figure out what function we're targeting.
	backingFunction := router.targetCache.BackingFunction(endpointID)
	if backingFunction == nil {
		retErr := errors.New("for endpoint ID:" + string(endpointID) + ", could not find backing function")
		return []byte{}, retErr
	}

	return router.CallFunction(*backingFunction, payload)
}

// CallFunction looks up a function and calls it.
func (router *Router) CallFunction(backingFunctionID functions.FunctionID, payload []byte) ([]byte, error) {
	backingFunction := router.targetCache.Function(backingFunctionID)
	if backingFunction == nil {
		resErr := errors.New("unable to look up backing function: " + string(backingFunctionID))
		return []byte{}, resErr
	}

	var fnGroup *functions.FunctionID
	var chosenFunction = backingFunction.ID

	if backingFunction.Group != nil {
		fnGroup = &backingFunction.ID
		chosen, err := backingFunction.Group.Functions.Choose()
		if err != nil {
			return []byte{}, err
		}
		chosenFunction = chosen
	}

	// Check if we need to send the input or output of the function to
	// any topics, and if so, we add that work to the work queue. The
	// work queue is a bounded channel, and we never block if it's full.
	// We need to be very loud though, so that we can autoscale up more
	// gateway instances ASAP.
	sendInput, sendOutput := router.subscribers(fnGroup, chosenFunction)

	if len(sendInput) > 0 {
		router.enqueueWork(sendInput, payload)
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

	// Check if we need to forward this to any topics enqueue the work
	// for forwarding events to subscribers, loudly dropping it if
	// we're congested.
	if len(sendOutput) > 0 {
		router.enqueueWork(sendOutput, result)
	}

	return result, nil
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
		router.work = make(chan work, router.NWorkers*2)
	}

	for i := 0; i < int(router.NWorkers); i++ {
		router.drainWaitGroup.Add(1)
		go router.Work()
	}
}

// Work is the main loop for a pub/sub worker goroutine
func (router *Router) Work() {
	for {
		// we use three select statements here to give preference
		// to the work chan, but fall-through to exiting when
		// the drain chan is closed and there's nothing to do.

		// 1. see if there's work in a non-blocking way
		select {
		case work := <-router.work:
			router.ProcessEvents(work)
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
			case work := <-router.work:
				router.ProcessEvents(work)
				continue
			default:
			}

			// no more work to do, decrement WaitGroup and return
			router.drainWaitGroup.Done()
			return
		case work := <-router.work:
			router.ProcessEvents(work)
		}
	}
}

// ProcessEvents sends events to a set of topics,
// and for each of the functions that get called
// as part of those topics, enqueue more work if
// they are producers for other topics.
func (router *Router) ProcessEvents(work work) {
	for _, topicID := range work.topics {
		subscribers := router.targetCache.SubscribersOfTopic(topicID)
		for _, subscriber := range subscribers {
			router.CallFunction(subscriber, work.payload)
		}
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
func (router *Router) WaitForEndpoint(endpointID endpoints.EndpointID) <-chan struct{} {
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

// WaitForPublisher returns a chan that is closed when a publisher is created.
// Primarily for testing purposes.
func (router *Router) WaitForFnPublisher(function functions.FunctionID, end string) <-chan struct{} {
	updatedChan := make(chan struct{})
	go func() {
		for {
			res := []pubsub.TopicID{}
			if end == "input" {
				res = router.targetCache.FunctionInputToTopics(function)
			} else if end == "output" {
				res = router.targetCache.FunctionOutputToTopics(function)
			} else {
				panic("WaitForFnPublisher received non-input/output function type.")
			}
			if len(res) > 0 {
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
