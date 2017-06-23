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
	TargetCache          targetcache.TargetCache
	dropMetric           prometheus.Counter
	log                  *zap.Logger
	NWorkers             uint
	drain                chan struct{}
	drainWaitGroup       sync.WaitGroup
	active               bool
	work                 chan work
	responseWriteTimeout time.Duration
}

type functionResponse struct {
	err     error
	payload []byte
}

// New instantiates a new Router
func New(TargetCache targetcache.TargetCache, dropMetric prometheus.Counter, log *zap.Logger) *Router {
	return &Router{
		TargetCache: TargetCache,
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

	resChan := make(chan functionResponse)

	go router.CallEndpoint(endpointID, reqBuf, resChan)

	if isAsync {
		w.WriteHeader(http.StatusAccepted)
	} else {
		res, ok := <-resChan
		if !ok {
			http.Error(w, "response channel unexpectedly closed", 500)
			return
		}

		if res.err != nil {
			http.Error(w, res.err.Error(), 500)
		} else {
			_, err := w.Write(res.payload)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
	}
}

func (router *Router) subscribers(fnGroup *functions.FunctionID, chosenFunction functions.FunctionID) (
	map[pubsub.TopicID]struct{}, map[pubsub.TopicID]struct{}, error) {

	sendInput := map[pubsub.TopicID]struct{}{}
	sendOutput := map[pubsub.TopicID]struct{}{}

	fillInputMap := func(fid functions.FunctionID) error {
		inputDest, err := router.TargetCache.FunctionInputToTopics(fid)
		if err != nil {
			return err
		}

		for _, topic := range inputDest {
			sendInput[topic] = struct{}{}
		}

		return nil
	}

	fillOutputMap := func(fid functions.FunctionID) error {
		outputDest, err := router.TargetCache.FunctionInputToTopics(fid)
		if err != nil {
			return err
		}

		for _, topic := range outputDest {
			sendOutput[topic] = struct{}{}
		}

		return nil
	}

	if fnGroup != nil {
		if err := fillInputMap(*fnGroup); err != nil {
			return nil, nil, err
		}

		if err := fillOutputMap(*fnGroup); err != nil {
			return nil, nil, err
		}
	}

	if err := fillInputMap(chosenFunction); err != nil {
		return nil, nil, err
	}

	if err := fillOutputMap(chosenFunction); err != nil {
		return nil, nil, err
	}

	return sendInput, sendOutput, nil
}

func (router *Router) enqueueWork(topicMap map[pubsub.TopicID]struct{}, payload []byte) {
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
func (router *Router) CallEndpoint(endpointID endpoints.EndpointID, payload []byte, resChan chan functionResponse) {
	res := functionResponse{}

	// 1. Figure out what function we're targeting

	backingFunction, err := router.TargetCache.BackingFunction(endpointID)
	if err != nil {
		res.err = errors.New("for endpoint ID:" + string(endpointID) + ", " + err.Error())
		resChan <- res
		return
	}

	var fnGroup *functions.FunctionID
	var chosenFunction = backingFunction.ID
	if backingFunction.Group != nil {
		fnGroup = &backingFunction.ID
		chosen, err := backingFunction.Group.Functions.Choose()
		if err != nil {
			res.err = errors.New("for endpoint ID:" + string(endpointID) + ", " + err.Error())
			resChan <- res
			return
		}
		chosenFunction = chosen
	}

	// 2. check if we need to send the input or output of the function to
	//    any topics, and if so, we add that work to the work queue. The
	//    work queue is a bounded channel, and we never block if it's full.
	//    We need to be very loud though, so that we can autoscale up more
	//    gateway instances ASAP

	sendInput, sendOutput, subscriberErr := router.subscribers(fnGroup, chosenFunction)
	if subscriberErr != nil {
		// We don't return because this is not fatal, and we still want to
		// call the function even though there's a problem with subscribers.
		res.err = errors.New("unable to determine subscribers for function: " + err.Error())
	} else {
		router.enqueueWork(sendInput, payload)
	}

	// 3. call the target backing function and submit the response to
	//		the resChan if it is not nil.
	result, err := router.CallFunction(chosenFunction, payload)
	if err != nil {
		res.err = errors.New("unable to reach backing function: " + err.Error())
		resChan <- res
		return
	}

	res.payload = result
	resChan <- res

	// 4. similar to #2, check if we need to forward this to any topics
	//		enqueue the work for forwarding events to subscribers, loudly
	//		dropping it if we're congested.

	if subscriberErr == nil {
		router.enqueueWork(sendOutput, result)
	}
}

// CallFunction looks up a function and calls it.
func (router *Router) CallFunction(fid functions.FunctionID, payload []byte) ([]byte, error) {
	f, err := router.TargetCache.Function(fid)
	if err != nil {
		return []byte{}, nil
	}
	return f.Call(payload)
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
