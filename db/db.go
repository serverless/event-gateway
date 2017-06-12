package db

import (
	"math"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

// eventType is used for distinguishing kinds of events.
type eventType uint

const (
	createdNode eventType = iota
	modifiedNode
	deletedNode
)

func init() {
	etcd.Register()
}

// CfgReactor is a type that can react to state changes on keys in a directory.
type CfgReactor interface {
	Created(key string, value []byte)
	Modified(key string, newValue []byte)
	Deleted(key string, lastKnownValue []byte)
}

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

// ReactiveCfgStore provides a means of watching for changes to
// interesting configuration in the backing database. It also
// maintains a cache of updates observed by a CfgReactor
// instance.
type ReactiveCfgStore struct {
	sync.RWMutex
	endpoints     string
	root          string
	cache         map[string][]byte
	kv            store.Store
	log           *zap.Logger
	backoffFactor int
}

// NewReactiveCfgStore instantiates a new ReactiveCfgStore.
func NewReactiveCfgStore(root string, endpoints []string, log *zap.Logger) *ReactiveCfgStore {
	if root == "/" {
		panic("Root (\"/\") used for watch path. Please namespace all usage to avoid performance issues.")
	} else if !strings.HasPrefix(root, "/") {
		panic("Path provided to NewReactiveCfgStore," + root + ", is invalid. Must begin with /")
	}

	kind := store.ETCD

	kv, err := libkv.NewStore(
		kind,
		endpoints,
		&store.Config{
			ConnectionTimeout: 10 * time.Second,
		},
	)

	if err != nil {
		log.Fatal("Cannot create kv client.",
			zap.Error(err))
	}

	noTrailingSlash := strings.TrimSuffix(root, "/")

	return &ReactiveCfgStore{
		root:          noTrailingSlash,
		endpoints:     strings.Join(endpoints, ","),
		cache:         map[string][]byte{},
		kv:            kv,
		log:           log,
		backoffFactor: 1,
	}
}

// React will watch for events on the ReactiveCfgStore's root directory,
// and call the Created/Modified/Deleted functions on a provided
// CfgReactor when changes are detected.
func (rfs *ReactiveCfgStore) React(reactor CfgReactor, shutdown chan struct{}) {
	events := make(chan event)
	go rfs.watchRoot(events, shutdown)

	go func() {
		for event := range events {
			switch event.eventType {
			case createdNode:
				reactor.Created(event.key, event.value)
			case modifiedNode:
				reactor.Modified(event.key, event.value)
			case deletedNode:
				reactor.Deleted(event.key, event.value)
			default:
				panic("received unknown event type.")
			}
		}
	}()
}

func (rfs *ReactiveCfgStore) bustCache() {
	rfs.Lock()
	defer rfs.Unlock()
	rfs.cache = map[string][]byte{}
}

func (rfs *ReactiveCfgStore) resetBackoff() {
	rfs.backoffFactor = 1
}

func (rfs *ReactiveCfgStore) backoff() {
	rfs.log.Warn("Backing-off after a failure",
		zap.String("event", "backoff"),
		zap.Int("seconds", rfs.backoffFactor))
	time.Sleep(time.Duration(rfs.backoffFactor) * time.Second)
	rfs.backoffFactor = int(math.Min(float64(rfs.backoffFactor<<1), 8))
}

func (rfs *ReactiveCfgStore) watchRoot(outgoingEvents chan event, shutdown chan struct{}) {
	// NB when extending libkv usage for DB's other than etcd, the watch behavior
	// will need to be carefully considered, as the code below will likely need
	// to be changed depending on which backend database is used.

	// This is the main loop for detecting changes.
	// 1. try to populate the watching root if it does not exist
	// 2. create a libkv watch chan for children of this root
	// 3. for each set of updates on this chan, diff it with
	//    our existing conception of the state, and emit
	//    event structs for each detected change.
	// 4. whenever we hit a problem, use a truncated
	//    exponential backoff to reduce load on any
	//    systems experiencing issues.
	for {
		// return if shutdown
		select {
		case <-shutdown:
			rfs.bustCache()
			return
		default:
		}

		// populate directory if it doesn't exist
		exists, err := rfs.Exists("")
		if err != nil {
			rfs.log.Error("Could not access database.",
				zap.String("event", "db"),
				zap.String("endpoints", rfs.endpoints),
				zap.String("key", rfs.root),
				zap.Error(err))
			rfs.backoff()
			continue
		}

		// TODO make sure it's a directory, not file

		if !exists {
			// must set IsDir to true since backend may be etcd
			err := rfs.Put("", []byte(""), &store.WriteOptions{IsDir: true})
			if err != nil {
				rfs.log.Error("Could not initialize watcher root.",
					zap.String("event", "db"),
					zap.String("endpoints", rfs.endpoints),
					zap.String("key", rfs.root),
					zap.Error(err))
				rfs.backoff()
				continue
			}
		}

		// create watch chan for this directory
		events, err := rfs.WatchTree("", shutdown)
		if err != nil {
			rfs.log.Error("Could not watch directory.",
				zap.String("event", "db"),
				zap.String("endpoints", rfs.endpoints),
				zap.String("key", rfs.root),
				zap.Error(err))
			rfs.backoff()
			continue
		}

		shouldShutdown := rfs.processEvents(events, outgoingEvents, shutdown)
		if shouldShutdown {
			return
		}
	}
}

func (rfs *ReactiveCfgStore) processEvents(incomingEvents <-chan []*store.KVPair,
	outgoingEvents chan event, shutdown chan struct{}) bool {

	cache := map[string]cachedValue{}
	for {
		select {
		case kvs, ok := <-incomingEvents:
			if ok {
				// compare all nodes against cache, emit diff events
				nextCache := rfs.diffCache(kvs, outgoingEvents, cache)

				// roll cache forward
				cache = nextCache

				// if we got here, all is well, so reset exponential backoff
				rfs.resetBackoff()
			} else {
				// directory nuked or connection to server failed
				rfs.log.Error("Either lost connection to db, or the watch path was deleted.",
					zap.String("event", "db"),
					zap.String("endpoints", rfs.endpoints),
					zap.String("key", rfs.root))

				rfs.backoff()
				return false
			}
		case <-shutdown:
			rfs.bustCache()
			return true
		}
	}
}

func (rfs *ReactiveCfgStore) diffCache(kvs []*store.KVPair, outgoingevents chan event,
	cache map[string]cachedValue) map[string]cachedValue {

	nextCache := map[string]cachedValue{}

	// Update all keys in the rfs cache because
	// we may have blown them away by busting cache
	// after shutting down a different watcher.
	rfs.Lock()
	for _, kv := range kvs {
		rfs.cache[kv.Key] = kv.Value
	}
	rfs.Unlock()

	for _, kv := range kvs {
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
	rfs.Lock()
	for key := range cache {
		delete(rfs.cache, key)
	}
	rfs.Unlock()

	for key, cachedValue := range cache {
		outgoingevents <- event{
			eventType: deletedNode,
			key:       key,
			value:     cachedValue.Value,
		}
	}

	return nextCache
}

// CachedGet looks for a key in the ReactiveCfgStore's cache,
// which is populated by events en-route to a CfgReactor.
// If the key is not found locally, a request is sent to the
// backing database.
func (rfs *ReactiveCfgStore) CachedGet(key string) ([]byte, error) {
	rfs.RLock()
	value, exists := rfs.cache[key]
	rfs.RUnlock()

	if exists {
		return value, nil
	}

	kv, err := rfs.kv.Get(key)

	// NB we don't put this into the cache because
	// we aren't watching its directory, if it
	// exists, and we would have a stale value
	// in cache as soon as it is updated.

	return kv.Value, err
}

// pass-through libkv API for simplifying access to the backing database.

// Put passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return rfs.kv.Put(rfs.root+"/"+key, value, options)
}

// Get passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) Get(key string) (*store.KVPair, error) {
	return rfs.kv.Get(rfs.root + "/" + key)
}

// Delete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) Delete(key string) error {
	return rfs.kv.Delete(rfs.root + "/" + key)
}

// Exists passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) Exists(key string) (bool, error) {
	return rfs.kv.Exists(rfs.root + "/" + key)
}

// Watch passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return rfs.kv.Watch(rfs.root+"/"+key, stopCh)
}

// WatchTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return rfs.kv.WatchTree(rfs.root+"/"+directory, stopCh)
}

// NewLock passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return rfs.kv.NewLock(rfs.root+"/"+key, options)
}

// List passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) List(directory string) ([]*store.KVPair, error) {
	return rfs.kv.List(rfs.root + "/" + directory)
}

// DeleteTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) DeleteTree(directory string) error {
	return rfs.kv.DeleteTree(rfs.root + "/" + directory)
}

// AtomicPut passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return rfs.kv.AtomicPut(rfs.root+"/"+key, value, previous, options)
}

// AtomicDelete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *ReactiveCfgStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return rfs.kv.AtomicDelete(rfs.root+"/"+key, previous)
}

// Close closes the underlying libkv client.
func (rfs *ReactiveCfgStore) Close() {
	rfs.kv.Close()
}
