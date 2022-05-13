package cache

import (
	"sync"
	"time"
)

const (
	NEVER_EXPIRED = 0
)

type Cache interface {
	Add(k interface{}, v Item)
	Get(k interface{}) (v Item, ok bool)
	Delete(k interface{})
}

type Item struct {
	value     interface{}
	timestamp time.Time
	expiring  time.Duration
}

func newCache() Cache {
	cache := inMemoryCache{
		items: make(map[interface{}]Item),
		mux:   new(sync.RWMutex),
	}
	return &cache
}

var globalCache Cache = newCache()

func Add(k string, v interface{}, expiring time.Duration) {
	globalCache.Add(k, Item{value: v, timestamp: time.Now(), expiring: expiring})
}

func Get(k string) (v interface{}, ok bool) {

	v, ok = globalCache.Get(k)
	if !ok {
		return nil, false
	}

	// expired, remove it and return nil
	if v.(Item).expiring != NEVER_EXPIRED && time.Now().After(v.(Item).timestamp.Add(v.(Item).expiring)) {
		globalCache.Delete(k)
		return nil, false
	}

	return v.(Item).value, true
}

func Delete(k string) {
	globalCache.Delete(k)
}

type inMemoryCache struct {
	items map[interface{}]Item
	mux   *sync.RWMutex
}

func (c *inMemoryCache) Add(k interface{}, v Item) {
	c.mux.Lock()
	c.items[k] = v
	c.mux.Unlock()
}

func (c *inMemoryCache) Get(k interface{}) (v Item, ok bool) {
	c.mux.RLock()
	v, ok = c.items[k]
	c.mux.RUnlock()
	return
}

func (c *inMemoryCache) Delete(k interface{}) {
	c.mux.Lock()
	delete(c.items, k)
	c.mux.Unlock()
	return
}
