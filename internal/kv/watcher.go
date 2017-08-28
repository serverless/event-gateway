package kv

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/serverless/libkv/store"
)

// Watcher provides a means of watching for changes to interesting configuration in the backing database. It also
// maintains a cache of updates observed by a Reactive instance.
type Watcher struct {
	path string
	kv   store.Store
	log  *zap.Logger

	// backoffFactor is used to track exponential backoffs
	// when failures occur.
	backoffFactor int

	// ReconciliationBaseDelay is the minimum duration in seconds
	// between re-connection attempts for db watches, which
	// mitigate connection issues causing lost updates.
	ReconciliationBaseDelay int

	// ReconciliationJitter is the maximum additional delay
	// applied to the ReconciliationBaseDelay when waiting
	// to reconnect to the db for Reconciliation
	ReconciliationJitter int
}

// NewWatcher instantiates a new Watcher.
func NewWatcher(path string, kv store.Store, log *zap.Logger) *Watcher {
	if path == "/" {
		panic("Root (\"/\") used for watch path. Please namespace all usage to avoid performance issues.")
	} else if !strings.HasPrefix(path, "/") {
		panic("Path provided to NewPathWatcher," + path + ", is invalid. Must begin with /")
	}
	// make sure we have a trailing slash for trimming future updates
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return &Watcher{
		path:                    path,
		kv:                      kv,
		log:                     log,
		backoffFactor:           1,
		ReconciliationBaseDelay: 30,
		ReconciliationJitter:    10,
	}
}

// React will watch for events on the PathWatcher's root directory,
// and call the Created/Modified/Deleted functions on a provided
// Reactive when changes are detected.
func (rfs *Watcher) React(reactor reactive, shutdown chan struct{}) {
	events := make(chan event)
	go rfs.watchRoot(events, shutdown)

	go func() {
		for event := range events {
			key := strings.TrimPrefix(event.key, rfs.path)
			switch event.eventType {
			case createdNode:
				reactor.Created(key, event.value)
			case modifiedNode:
				reactor.Modified(key, event.value)
			case deletedNode:
				reactor.Deleted(key, event.value)
			default:
				panic("Received unknown event type.")
			}
		}
	}()
}

const (
	createdNode eventType = iota
	modifiedNode
	deletedNode
)

// eventType is used for distinguishing kinds of events.
type eventType uint

type event struct {
	eventType eventType
	key       string

	// Value is set to the new value for CREATED/MODIFIED events,
	// and it is set to the last known value for DELETED events.
	value []byte
}

type cachedValue struct {
	Value     []byte
	LastIndex uint64
}

func (rfs *Watcher) resetBackoff() {
	rfs.backoffFactor = 1
}

func (rfs *Watcher) backoff() {
	rfs.log.Info("Backing-off after a failure",
		zap.String("event", "backoff"),
		zap.Int("seconds", rfs.backoffFactor))
	time.Sleep(time.Duration(rfs.backoffFactor) * time.Second)
	rfs.backoffFactor = int(math.Min(float64(rfs.backoffFactor<<1), 8))
}

func (rfs *Watcher) reconciliationTimeout() <-chan time.Time {
	// use a minimum jitter of 1
	maxJitter := int(math.Max(float64(rfs.ReconciliationJitter), 1))
	jitter := rand.Intn(maxJitter)
	delay := time.Duration(jitter+rfs.ReconciliationBaseDelay) * time.Second
	return time.After(delay)
}

func (rfs *Watcher) watchRoot(outgoingEvents chan event, shutdown chan struct{}) {
	// This is the main loop for detecting changes.
	// 1. try to populate the watching root if it does not exist
	// 2. create a libkv watch chan for children of this root
	// 3. for each set of updates on this chan, diff it with
	//    our existing conception of the state, and emit
	//    event structs for each detected change.
	// 4. whenever we hit a problem, use a truncated
	//    exponential backoff to reduce load on any
	//    systems experiencing issues.

	cache := map[string]cachedValue{}

	for {
		// return if shutdown
		select {
		case <-shutdown:
			return
		default:
		}

		// populate directory if it doesn't exist
		exists, err := rfs.kv.Exists(rfs.path)
		if err != nil {
			rfs.log.Error("Could not access database.",
				zap.String("event", "db"),
				zap.String("key", rfs.path),
				zap.Error(err))
			rfs.backoff()
			continue
		}

		if !exists {
			err = rfs.kv.Put(rfs.path, []byte(nil), nil)
			if err != nil {
				if strings.HasPrefix(err.Error(), "102: Not a file") {
					rfs.log.Debug("Another node (probably) created the root directory first.")
				} else {
					rfs.log.Error("Could not initialize watcher root.",
						zap.String("event", "db"),
						zap.String("key", rfs.path),
						zap.Error(err))
					rfs.backoff()
					continue
				}
			}
		}

		// create watch chan for this directory
		events, err := rfs.kv.WatchTree(rfs.path, shutdown)
		if err != nil {
			rfs.log.Error("Could not watch directory.",
				zap.String("event", "db"),
				zap.String("key", rfs.path),
				zap.Error(err))
			rfs.backoff()
			continue
		}

		// process events from the events chan until the
		// connection to the server is lost, or the key
		// is removed.
		shouldShutdown := rfs.processEvents(&cache, events, outgoingEvents, shutdown)
		if shouldShutdown {
			return
		}
	}
}

func (rfs *Watcher) processEvents(cache *map[string]cachedValue, incomingEvents <-chan []*store.KVPair,
	outgoingEvents chan event, shutdown chan struct{}) bool {

	for {
		select {
		case kvs, ok := <-incomingEvents:
			if ok {
				// compare all nodes against cache, emit diff events
				nextCache := rfs.diffCache(kvs, outgoingEvents, *cache)

				// roll cache forward
				*cache = nextCache

				// if we got here, all is well, so reset exponential backoff
				rfs.resetBackoff()
			} else {
				// directory nuked or connection to server failed
				rfs.log.Error("Either lost connection to db, or the watch path was deleted.",
					zap.String("event", "db"),
					zap.String("key", rfs.path))

				rfs.backoff()
				return false
			}
		case <-shutdown:
			return true
		}
	}
}

func (rfs *Watcher) diffCache(kvs []*store.KVPair, outgoingevents chan event,
	cache map[string]cachedValue) map[string]cachedValue {

	nextCache := map[string]cachedValue{}

	for _, kv := range kvs {
		// Is directory
		if kv.Value == nil {
			continue
		}

		// populate next cache
		nextCache[kv.Key] = cachedValue{
			Value:     kv.Value,
			LastIndex: kv.LastIndex,
		}

		old, exists := cache[kv.Key]
		if exists {
			// if LastIndex newer, emit modifiedNode event
			if kv.LastIndex > old.LastIndex {
				outgoingevents <- event{
					eventType: modifiedNode,
					key:       kv.Key,
					value:     kv.Value,
				}
			}

			// Remove key from old cache so we can
			// learn about any deleted nodes.
			delete(cache, kv.Key)
		} else {
			// this node wasn't present before, emit createdNode
			outgoingevents <- event{
				eventType: createdNode,
				key:       kv.Key,
				value:     kv.Value,
			}
		}
	}

	// Anything that was present in the old cache,
	// but not in our recent update, has been deleted
	// on the database. Clear it from our cache and
	// emit a deletedNode event.
	for key, cachedValue := range cache {
		outgoingevents <- event{
			eventType: deletedNode,
			key:       key,
			value:     cachedValue.Value,
		}
	}

	return nextCache
}

// Reactive is a type that can react to state changes on keys in a directory.
type reactive interface {
	Created(key string, value []byte)
	Modified(key string, newValue []byte)
	Deleted(key string, lastKnownValue []byte)
}
