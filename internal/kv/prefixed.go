package kv

import (
	"strings"

	"github.com/serverless/libkv/store"
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
func (ps *PrefixedStore) Put(key string, value []byte, options *store.WriteOptions) error {
	return ps.kv.Put(ps.root+key, value, options)
}

// Get passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) Get(key string, options *store.ReadOptions) (*store.KVPair, error) {
	return ps.kv.Get(ps.root+key, options)
}

// Delete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) Delete(key string) error {
	return ps.kv.Delete(ps.root + key)
}

// Exists passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) Exists(key string, options *store.ReadOptions) (bool, error) {
	return ps.kv.Exists(ps.root+key, options)
}

// Watch passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) Watch(key string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan *store.KVPair, error) {
	return ps.kv.Watch(ps.root+key, stopCh, options)
}

// WatchTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) WatchTree(directory string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan []*store.KVPair, error) {
	return ps.kv.WatchTree(ps.root+directory, stopCh, options)
}

// NewLock passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return ps.kv.NewLock(ps.root+key, options)
}

// List passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) List(directory string, options *store.ReadOptions) ([]*store.KVPair, error) {
	prefixed, err := ps.kv.List(ps.root+directory, options)
	if err != nil {
		return nil, err
	}

	unprefixed := []*store.KVPair{}
	for _, kv := range prefixed {
		// Is directory
		if kv.Key == ps.root {
			continue
		}

		unprefixed = append(unprefixed, &store.KVPair{
			Key:       strings.TrimPrefix(kv.Key, ps.root),
			Value:     kv.Value,
			LastIndex: kv.LastIndex,
		})
	}

	return unprefixed, nil
}

// DeleteTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) DeleteTree(directory string) error {
	return ps.kv.DeleteTree(ps.root + directory)
}

// AtomicPut passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return ps.kv.AtomicPut(ps.root+key, value, previous, options)
}

// AtomicDelete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *PrefixedStore) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return ps.kv.AtomicDelete(ps.root+key, previous)
}

// Close closes the underlying libkv client.
func (ps *PrefixedStore) Close() {
	ps.kv.Close()
}
