package kv

import (
	"strings"

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
		path:          path,
		kv:            kv,
		log:           log,
		backoffFactor: 1,
	}
}

// React will watch for events on the PathWatcher's root directory,
// and call the Created/Modified/Deleted functions on a provided
// Reactive when changes are detected.
func (rfs *Watcher) React(reactor Reactor, shutdown chan struct{}) {
	events := make(chan event)
	go rfs.watchRoot(events, shutdown)

	go func() {
		for event := range events {
			key := strings.TrimPrefix(event.key, rfs.path)
			switch event.eventType {
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
	modifiedNode eventType = iota
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

func (rfs *Watcher) watchRoot(outgoingEvents chan event, shutdown chan struct{}) {
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
			continue
		}

		// process events from the events chan until the
		// connection to the server is lost, or the key
		// is removed.
		shouldShutdown := rfs.processEvents(events, outgoingEvents, shutdown)
		if shouldShutdown {
			return
		}
	}
}

func (rfs *Watcher) processEvents(incoming <-chan []*store.KVPair, outgoing chan event, shutdown chan struct{}) bool {
	cache := map[string]cachedValue{}

	for {
		select {
		case kvs, ok := <-incoming:
			if ok {
				for _, kv := range kvs {
					// Is directory
					if kv.Key == rfs.path {
						continue
					}

					if kv.Value != nil {
						outgoing <- event{
							eventType: modifiedNode,
							key:       kv.Key,
							value:     kv.Value,
						}

						cache[kv.Key] = cachedValue{
							Value:     kv.Value,
							LastIndex: kv.LastIndex,
						}
					} else {
						old, exists := cache[kv.Key]
						if exists {
							outgoing <- event{
								eventType: deletedNode,
								key:       kv.Key,
								value:     old.Value,
							}

							delete(cache, kv.Key)
						}
					}
				}
			} else {
				// directory nuked or connection to server failed
				rfs.log.Error("Either lost connection to db, or the watch path was deleted.",
					zap.String("event", "db"),
					zap.String("key", rfs.path))
				return false
			}
		case <-shutdown:
			return true
		}
	}
}
