package router

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/serverless/gateway/endpoints"
	"github.com/serverless/gateway/targetcache"
)

type Router struct {
	targetCache   targetcache.TargetCache
	log           *zap.Logger
	NWorkers      uint
	drain         chan struct{}
	drainComplete chan struct{}
	work          chan work
}

type res struct {
	e       error
	payload []byte
}

func New(targetCache targetcache.TargetCache, log *zap.Logger) *Router {
	return &Router{
		targetCache:   targetCache,
		log:           log,
		NWorkers:      20,
		drain:         make(chan struct{}),
		drainComplete: make(chan struct{}),
		work:          nil,
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
			fmt.Fprintf(w, res.payload)
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
	backingFunctions, err := router.targetCache.BackingFunctions(endpointID)

	// 2. check if we need to send the input of the function to any topics,
	//    and if so, we add that work to the work queue. The work queue is
	//		a bounded channel, and we never block if it's full. We need to
	//		be very loud though, so that we can autoscale up more gateway
	//		instances ASAP

	// 3. call the target backing function and submit the response to
	//		the resChan if it is not nil.
	router.CallFunction(fid, payload)

	// 4. similar to #2, check if we need to forward this to any topics
	//		(functions subscribing to topics) and enqueue the work,
	//		loudly dropping it if we're congested.

}

func (router *Router) CallFunction(fid functions.FunctionID, payload []byte) {

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
