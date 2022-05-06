package cache

import (
	"sync"
	"time"
)

const (
	NEVER_EXPIRED = 0
)

type item struct {
	value     interface{}
	timestamp time.Time
	expiring  time.Duration
}

type cache struct {
	items map[string]item
}

var c *cache = newCache()
var mux *sync.RWMutex = new(sync.RWMutex)

func newCache() *cache {
	cache := cache{
		items: make(map[string]item),
	}
	return &cache
}

func Add(k string, v interface{}, expiring time.Duration) {
	mux.Lock()
	c.items[k] = item{value: v, timestamp: time.Now(), expiring: expiring}
	mux.Unlock()
}

func Get(k string) (v interface{}, ok bool) {
	mux.RLock()
	defer mux.RUnlock()

	item, ok := c.items[k]

	// expired, remove it and return nil
	if item.expiring != NEVER_EXPIRED && time.Now().After(item.timestamp.Add(item.expiring)) {
		delete(c.items, k)
		return nil, false
	}

	return item.value, ok
}

func Delete(k string) {
	mux.Lock()
	delete(c.items, k)
	mux.Unlock()
}
