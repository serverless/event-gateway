package kv

// Reactor is a type that can react to state changes on keys in a directory.
type Reactor interface {
	Modified(key string, value []byte)
	Deleted(key string, value []byte)
}
