package cache

import (
	"time"
)

const (
	NEVER_EXPIRED = 0

	IN_MEMORY = "InMemory"
)

type cacheNewFunc func() Cache

var CacheNewFuncs = make(map[string]cacheNewFunc)

func init() {
	CacheNewFuncs[IN_MEMORY] = NewInMemoryMap
}

type Cache interface {
	Add(k interface{}, v interface{}, expiring time.Duration)
	Get(k interface{}) (v interface{}, ok bool)
	Delete(k interface{})
}

var globalCache Cache

// New returns t type of cache.
// It will return InMemoryCache if t type is not available.
func New(t string) Cache {
	f, ok := CacheNewFuncs[t]
	if !ok {
		return CacheNewFuncs[IN_MEMORY]()
	}
	return f()
}

// Add adds item to global cache
func Add(k string, v interface{}, expiring time.Duration) {
	if globalCache == nil {
		globalCache = New(IN_MEMORY)
	}

	globalCache.Add(k, v, expiring)
}

// Get gets item from global cache
func Get(k string) (v interface{}, ok bool) {
	if globalCache == nil {
		globalCache = New(IN_MEMORY)
	}

	return globalCache.Get(k)
}

// Delete deletes item in global cache
func Delete(k string) {
	if globalCache == nil {
		globalCache = New(IN_MEMORY)
	}

	globalCache.Delete(k)
}
