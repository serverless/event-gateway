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
}

func New(targetCache targetcache.TargetCache, log *zap.Logger) *Router {
	return &Router{
		targetCache:   targetCache,
		log:           log,
		NWorkers:      20,
		drain:         make(chan struct{}),
		drainComplete: make(chan struct{}),
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

	id := strings.ToLower(r.Method) + "-" + r.URL.EscapedPath()
	endpointID := endpoints.EndpointID(id)
	router.log.Info("got a new request: " + string(endpointID))
	backingFunctions, err := router.targetCache.BackingFunctions(endpointID)
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
