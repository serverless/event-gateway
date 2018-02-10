package store

import (
	"strings"

	"github.com/serverless/libkv/store"
)

// Prefixed namespaces a libkv Store instance.
type Prefixed struct {
	root string
	kv   store.Store
}

// NewPrefixed creates a new namespaced libkv Store.
func NewPrefixed(root string, kv store.Store) *Prefixed {
	if !strings.HasSuffix(root, "/") {
		root = root + "/"
	}
	return &Prefixed{
		root: root,
		kv:   kv,
	}
}

// Put passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) Put(key string, value []byte, options *store.WriteOptions) error {
	return ps.kv.Put(ps.root+key, value, options)
}

// Get passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) Get(key string, options *store.ReadOptions) (*store.KVPair, error) {
	return ps.kv.Get(ps.root+key, options)
}

// Delete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) Delete(key string) error {
	return ps.kv.Delete(ps.root + key)
}

// Exists passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) Exists(key string, options *store.ReadOptions) (bool, error) {
	return ps.kv.Exists(ps.root+key, options)
}

// Watch passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) Watch(key string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan *store.KVPair, error) {
	return ps.kv.Watch(ps.root+key, stopCh, options)
}

// WatchTree passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) WatchTree(directory string, stopCh <-chan struct{}, options *store.ReadOptions) (<-chan []*store.KVPair, error) {
	return ps.kv.WatchTree(ps.root+directory, stopCh, options)
}

// NewLock passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return ps.kv.NewLock(ps.root+key, options)
}

// List passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) List(directory string, options *store.ReadOptions) ([]*store.KVPair, error) {
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
func (ps *Prefixed) DeleteTree(directory string) error {
	return ps.kv.DeleteTree(ps.root + directory)
}

// AtomicPut passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return ps.kv.AtomicPut(ps.root+key, value, previous, options)
}

// AtomicDelete passes requests to the underlying libkv implementation, appending the root to paths for isolation.
func (ps *Prefixed) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return ps.kv.AtomicDelete(ps.root+key, previous)
}

// Close closes the underlying libkv client.
func (ps *Prefixed) Close() {
	ps.kv.Close()
}
