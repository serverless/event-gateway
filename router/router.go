package router

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
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

type Router struct {
	sync.Mutex
	targetCache          targetcache.TargetCache
	dropMetric           prometheus.Counter
	log                  *zap.Logger
	NWorkers             uint
	drain                chan struct{}
	drainComplete        chan struct{}
	work                 chan work
	responseWriteTimeout time.Duration
}

type res struct {
	e       error
	payload []byte
}

func New(targetCache targetcache.TargetCache, dropMetric prometheus.Counter, log *zap.Logger) *Router {
	return &Router{
		targetCache:          targetCache,
		dropMetric:           dropMetric,
		log:                  log,
		NWorkers:             20,
		drain:                make(chan struct{}),
		drainComplete:        make(chan struct{}),
		work:                 nil,
		responseWriteTimeout: 3 * time.Second,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if we're draining requests, spit back a 503
	select {
	case <-router.drain:
		http.Error(w, http.StatusText(503), 503)
		return
	default:
	}

	rawPath := r.URL.EscapedPath()
	isAsync := strings.HasSuffix(rawPath, "/async")
	trimmedPath := strings.TrimSuffix(rawPath, "/async")
	id := strings.ToLower(r.Method) + "-" + trimmedPath
	endpointID := endpoints.EndpointID(id)
	router.log.Debug("got a new request: " + string(endpointID))

	reqBuf, err := ioutil.ReadAll(r.Body)

	if err != nil {
		fmt.Fprintf(w, "%s", err)
	}

	resChan := make(chan res)

	go router.CallEndpoint(endpointID, reqBuf, resChan)

	if isAsync {
		w.WriteHeader(http.StatusAccepted)
	} else {
		res, ok := <-resChan
		if !ok {
			http.Error(w, "response channel unexpectedly closed", 500)
			return
		}
		if res.e != nil {
			http.Error(w, res.e.Error(), 500)
		} else {
			timeout := time.After(router.responseWriteTimeout)
			total := 0
			sz := len(res.payload)
			for total < sz {
				select {
				case <-timeout:
					http.Error(w, "timed out writing response", 500)
					return
				default:
				}

				written, err := w.Write(res.payload)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				total += written
			}
		}
	}
}

func (router *Router) CallEndpoint(endpointID endpoints.EndpointID, payload []byte, resChan chan res) {
	// 1. Get the backing functions, note that we need to change the
	// BackingFunctions to let us know if it was a group, so we can
	// submit events to topics based both on the function group AND
	// the chosen function if subscribers are different for them, but
	// being careful to not send duplicate events if both are registered
	// as producers for a topic.
	sendInput := map[pubsub.TopicID]struct{}{}
	sendOutput := map[pubsub.TopicID]struct{}{}
	res := res{}

	backingFunctions, fnGroup, err := router.targetCache.BackingFunctions(endpointID)

	var chosenFunction functions.FunctionID

	if len(backingFunctions) == 1 {
		chosenFunction = backingFunctions[0].FunctionID
	} else {
		weightTotal := uint(0)
		for _, wf := range backingFunctions {
			weightTotal += wf.Weight
		}

		if weightTotal < 1 {
			res.e = errors.New("for endpoint ID:" + string(endpointID) +
				" the target function weights sum to 0, there is not one function to target.")
			resChan <- res
			return
		}

		chosenWeight := uint(1 + rand.Intn(int(weightTotal)))
		weightsSoFar := uint(0)
		for _, wf := range backingFunctions {
			chosenFunction = wf.FunctionID
			weightsSoFar += wf.Weight
			if weightsSoFar >= chosenWeight {
				break
			}
		}
	}

	// 2. check if we need to send the input or output of the function to
	//    any topics, and if so, we add that work to the work queue. The
	//		work queue is	a bounded channel, and we never block if it's full.
	//		We need to be very loud though, so that we can autoscale up more
	//		gateway instances ASAP

	fillInputMap := func(fid functions.FunctionID) bool {
		inputDest, err := router.targetCache.FunctionInputToTopics(fid)
		if err != nil {
			res.e = errors.New("Unable to find subscribing topics for the input to function: " + string(fid))
			resChan <- res
			return false
		}

		for _, topic := range inputDest {
			sendInput[topic] = struct{}{}
		}

		return true
	}

	fillOutputMap := func(fid functions.FunctionID) bool {
		outputDest, err := router.targetCache.FunctionInputToTopics(fid)
		if err != nil {
			res.e = errors.New("Unable to find subscribing topics for the output to function: " + string(fid))
			resChan <- res
			return false
		}

		for _, topic := range outputDest {
			sendOutput[topic] = struct{}{}
		}

		return true
	}

	if fnGroup != nil {
		if !fillInputMap(*fnGroup) {
			return
		}

		if !fillOutputMap(*fnGroup) {
			return
		}
	}

	if !fillInputMap(chosenFunction) {
		return
	}

	if !fillOutputMap(chosenFunction) {
		return
	}

	for topic, _ := range sendInput {
		select {
		case router.work <- work{
			topic:   topic,
			payload: payload,
		}:
		default:
			// We could not submit any work, this is NOT good but
			// we will sacrifice consistency for availability for now.
			router.dropMetric.Inc()
		}
	}

	// 3. call the target backing function and submit the response to
	//		the resChan if it is not nil.
	result, err := router.CallFunction(chosenFunction, payload)
	if err != nil {
		res.e = errors.New("unable to reach backing function: " + err.Error())
		resChan <- res
		return
	}

	res.payload = result
	resChan <- res

	// 4. similar to #2, check if we need to forward this to any topics
	//		enqueue the work for forwarding events to subscribers, loudly
	//		dropping it if we're congested.

	for topic, _ := range sendOutput {
		select {
		case router.work <- work{
			topic:   topic,
			payload: result,
		}:
		default:
			// We could not submit any work, this is NOT good but
			// we will sacrifice consistency for availability for now.
			router.dropMetric.Inc()
		}
	}
}

func (router *Router) CallFunction(fid functions.FunctionID, payload []byte) ([]byte, error) {
	return []byte{}, nil
}

// StartWorkers spins up NWorkers goroutines for processing
// the event subscriptions.
func (router *Router) StartWorkers() {}

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
	<-router.drainComplete
}
