package cache

import (
	"time"
)

const (
	NEVER_EXPIRED = 0
)

type cacheNewFunc func() Cache

var CacheNewFuncs = make(map[string]cacheNewFunc)

func init() {
	CacheNewFuncs["InMemory"] = NewInMemoryMap
	CacheNewFuncs["InMemoryMutux"] = NewInMemoryMutex
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
		return CacheNewFuncs["InMemory"]()
	}
	return f()
}

// Add adds item to global cache
func Add(k string, v interface{}, expiring time.Duration) {
	if globalCache == nil {
		globalCache = New("InMemory")
	}

	globalCache.Add(k, v, expiring)
}

// Get gets item from global cache
func Get(k string) (v interface{}, ok bool) {
	if globalCache == nil {
		globalCache = New("InMemory")
	}

	return globalCache.Get(k)
}

// Delete deletes item in global cache
func Delete(k string) {
	if globalCache == nil {
		globalCache = New("InMemory")
	}

	globalCache.Delete(k)
}
