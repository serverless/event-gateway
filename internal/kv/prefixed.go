package kv

import (
	"strings"

	"github.com/docker/libkv/store"
)

// PrefixedStore namespaces a libkv Store instance.
type PrefixedStore struct {
	root string
	kv   store.Store
}

// NewPrefixedStore creates a new namespaced libkv Store.
func NewPrefixedStore(root string, kv store.Store) *PrefixedStore {
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}
	return &PrefixedStore{
		root: root,
		kv:   kv,
	}
}

// Put passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return rfs.kv.Put(rfs.root+key, value, options)
}

// Get passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) Get(key string) (*store.KVPair, error) {
	return rfs.kv.Get(rfs.root + key)
}

// Delete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) Delete(key string) error {
	return rfs.kv.Delete(rfs.root + key)
}

// Exists passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) Exists(key string) (bool, error) {
	return rfs.kv.Exists(rfs.root + key)
}

// Watch passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return rfs.kv.Watch(rfs.root+key, stopCh)
}

// WatchTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return rfs.kv.WatchTree(rfs.root+directory, stopCh)
}

// NewLock passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return rfs.kv.NewLock(rfs.root+key, options)
}

// List passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) List(directory string) ([]*store.KVPair, error) {
	prefixed, err := rfs.kv.List(rfs.root + directory)
	if err != nil {
		return nil, err
	}

	unprefixed := []*store.KVPair{}
	for _, kv := range prefixed {
		unprefixed = append(unprefixed, &store.KVPair{
			Key:       strings.TrimPrefix(kv.Key, rfs.root),
			Value:     kv.Value,
			LastIndex: kv.LastIndex,
		})
	}

	return unprefixed, nil
}

// DeleteTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) DeleteTree(directory string) error {
	return rfs.kv.DeleteTree(rfs.root + directory)
}

// AtomicPut passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return rfs.kv.AtomicPut(rfs.root+key, value, previous, options)
}

// AtomicDelete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (rfs *PrefixedStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return rfs.kv.AtomicDelete(rfs.root+key, previous)
}

// Close closes the underlying libkv client.
func (rfs *PrefixedStore) Close() {
	rfs.kv.Close()
}
