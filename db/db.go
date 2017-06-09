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

type EventType uint

const (
	CREATED_NODE EventType = iota
	MODIFIED_NODE
	DELETED_NODE
)

func init() {
	etcd.Register()
	// TODO for other stores, do consul.Register(), zookeeper.Register() also
}

type CfgReactor interface {
	Created(key string, value []byte)
	Modified(key string, newValue []byte)
	Deleted(key string, lastKnownValue []byte)
}

type Event struct {
	EventType EventType
	Key       string

	// Value is set to the new value for CREATED/MODIFIED events,
	// and it is set to the last known value for DELETED events.
	Value []byte
}

type CachedValue struct {
	Value     []byte
	LastIndex uint64
}

type ReactiveCfgStore struct {
	sync.RWMutex
	endpoints string
	root      string
	cache     map[string][]byte
	kv        store.Store
	log       *zap.Logger
}

func NewReactiveCfgStore(root, kindStr string, endpoints []string, log *zap.Logger) *ReactiveCfgStore {
	kind := store.ETCD
	if kindStr != "etcd" {
		panic("--db-type set to unsupported value: " + kindStr)
	}

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

	return &ReactiveCfgStore{
		root:      root,
		endpoints: strings.Join(endpoints, ","),
		cache:     map[string][]byte{},
		kv:        kv,
		log:       log,
	}
}

func (rfs *ReactiveCfgStore) React(reactor CfgReactor, shutdown chan struct{}) {
	events := rfs.watchRoot(shutdown)

	go func() {
		for event := range events {
			switch event.EventType {
			case CREATED_NODE:
				reactor.Created(event.Key, event.Value)
			case MODIFIED_NODE:
				reactor.Modified(event.Key, event.Value)
			case DELETED_NODE:
				reactor.Deleted(event.Key, event.Value)
			default:
				panic("received unknown event type.")
			}
		}
	}()
}

func (rfs *ReactiveCfgStore) watchRoot(shutdown chan struct{}) chan Event {
	outgoingEvents := make(chan Event)

	// TODO this is only tested for etcd, may need to change logic
	// for zk/consul cache allows us to diff nodes we receive
	go func() {
		cache := map[string]CachedValue{}

		backoffFactor := 1
		backoff := func() {
			rfs.log.Warn("Backing-off after a failure",
				zap.String("event", "backoff"),
				zap.Int("seconds", backoffFactor))
			time.Sleep(time.Duration(backoffFactor) * time.Second)
			backoffFactor = int(math.Min(float64(backoffFactor<<1), 8))
		}
		resetBackoff := func() {
			backoffFactor = 1
		}

		for {
			// populate directory if it doesn't exist
			exists, err := rfs.Exists("")
			if err != nil {
				rfs.log.Error("Could not access database.",
					zap.String("event", "db"),
					zap.String("endpoints", rfs.endpoints),
					zap.String("key", rfs.root),
					zap.Error(err))
				backoff()
				continue
			}
			if !exists {
				// must set IsDir to true since backend may be etcd
				err := rfs.Put("", []byte(""), &store.WriteOptions{IsDir: true})
				if err != nil {
					rfs.log.Error("Could not initialize watcher root.",
						zap.String("event", "db"),
						zap.String("endpoints", rfs.endpoints),
						zap.String("key", rfs.root),
						zap.Error(err))
					backoff()
					continue
				}
			}

			events, err := rfs.WatchTree("", shutdown)
			if err != nil {
				rfs.log.Error("Could not watch directory.",
					zap.String("event", "db"),
					zap.String("endpoints", rfs.endpoints),
					zap.String("key", rfs.root),
					zap.Error(err))
				backoff()
				continue
			}

		innerLoop:
			for {
				select {
				case kvs, ok := <-events:
					if ok {
						// compare all nodes against cache, emit diff events
						nextCache := map[string]CachedValue{}
						for _, kv := range kvs {
							// populate next cache
							nextCache[kv.Key] = CachedValue{
								Value:     kv.Value,
								LastIndex: kv.LastIndex,
							}

							old, exists := cache[kv.Key]
							if exists {
								// if LastIndex newer, emit MODIFIED_NODE event
								if kv.LastIndex > old.LastIndex {
									rfs.Lock()
									rfs.cache[kv.Key] = kv.Value
									rfs.Unlock()
									outgoingEvents <- Event{
										EventType: MODIFIED_NODE,
										Key:       kv.Key,
										Value:     kv.Value,
									}
								}

								// Remove key from old cache so we can
								// learn about any deleted nodes.
								delete(cache, kv.Key)
							} else {
								// this node wasn't present before, emit CREATED_NODE
								rfs.Lock()
								rfs.cache[kv.Key] = kv.Value
								rfs.Unlock()
								outgoingEvents <- Event{
									EventType: CREATED_NODE,
									Key:       kv.Key,
									Value:     kv.Value,
								}
							}
						}

						for key, cachedValue := range cache {
							// this node is not present anymore, emit
							// DELETED_NODE event.
							rfs.Lock()
							delete(rfs.cache, key)
							rfs.Unlock()
							outgoingEvents <- Event{
								EventType: DELETED_NODE,
								Key:       key,
								Value:     cachedValue.Value,
							}
						}

						// roll cache forward
						cache = nextCache
						resetBackoff()
					} else {
						// directory nuked or connection to server failed
						rfs.log.Error("Either lost connection to db, or the watch path was deleted.",
							zap.String("event", "db"),
							zap.String("endpoints", rfs.endpoints),
							zap.String("key", rfs.root))

						backoff()
						break innerLoop
					}
				case <-shutdown:
					return
				}
			}
		}
	}()
	return outgoingEvents
}

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
func (rfs *ReactiveCfgStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return rfs.kv.Put(rfs.root+"/"+key, value, options)
}
func (rfs *ReactiveCfgStore) Get(key string) (*store.KVPair, error) {
	return rfs.kv.Get(rfs.root + "/" + key)
}
func (rfs *ReactiveCfgStore) Delete(key string) error {
	return rfs.kv.Delete(rfs.root + "/" + key)
}
func (rfs *ReactiveCfgStore) Exists(key string) (bool, error) {
	return rfs.kv.Exists(rfs.root + "/" + key)
}
func (rfs *ReactiveCfgStore) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return rfs.kv.Watch(rfs.root+"/"+key, stopCh)
}
func (rfs *ReactiveCfgStore) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return rfs.kv.WatchTree(rfs.root+"/"+directory, stopCh)
}
func (rfs *ReactiveCfgStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return rfs.kv.NewLock(rfs.root+"/"+key, options)
}
func (rfs *ReactiveCfgStore) List(directory string) ([]*store.KVPair, error) {
	return rfs.kv.List(rfs.root + "/" + directory)
}
func (rfs *ReactiveCfgStore) DeleteTree(directory string) error {
	return rfs.kv.DeleteTree(rfs.root + "/" + directory)
}
func (rfs *ReactiveCfgStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return rfs.kv.AtomicPut(rfs.root+"/"+key, value, previous, options)
}
func (rfs *ReactiveCfgStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return rfs.kv.AtomicDelete(rfs.root+"/"+key, previous)
}
func (rfs *ReactiveCfgStore) Close() {
	rfs.kv.Close()
}
