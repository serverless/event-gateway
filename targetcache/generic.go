package targetcache

type cacheMaintainer struct {
	setter  func(string, []byte)
	deleter func(string, []byte)
}

func newCacheMaintainer(setter func(string, []byte), deleter func(string, []byte)) *cacheMaintainer {
	return &cacheMaintainer{
		setter:  setter,
		deleter: deleter,
	}
}

// Created is called when a new endpoint is detected in the config.
func (c *cacheMaintainer) Created(key string, value []byte) {
	c.setter(key, value)
}

// Modified is called when an existing endpoint is modified in the config.
func (c *cacheMaintainer) Modified(key string, newValue []byte) {
	c.setter(key, newValue)
}

// Deleted is called when a endpoint is deleted in the config.
func (c *cacheMaintainer) Deleted(key string, lastKnownValue []byte) {
	c.deleter(key, lastKnownValue)
}
