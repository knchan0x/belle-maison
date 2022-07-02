package cache

import (
	"sync"
	"time"
)

type InMemoryMap struct {
	items sync.Map
}

type Item struct {
	value     interface{}
	timestamp time.Time
	expiring  time.Duration
}

func NewInMemoryMap() Cache {
	return &InMemoryMap{}
}

func (c *InMemoryMap) Add(k interface{}, v interface{}, expiring time.Duration) {
	c.items.Store(k, Item{value: v, timestamp: time.Now(), expiring: expiring})
}

func (c *InMemoryMap) Get(k interface{}) (interface{}, bool) {
	v, ok := c.items.Load(k)

	// not exists or expired, remove it and return nil if expired
	if !ok || v.(Item).expiring != NEVER_EXPIRED && time.Now().After(v.(Item).timestamp.Add(v.(Item).expiring)) {
		c.Delete(k)
		return nil, false
	}

	return v.(Item).value, true
}

func (c *InMemoryMap) Delete(k interface{}) {
	c.items.Delete(k)
	return
}

type InMemoryMutex struct {
	items map[interface{}]Item
	mux   sync.RWMutex
}

func NewInMemoryMutex() Cache {
	return &InMemoryMutex{
		items: make(map[interface{}]Item),
		mux:   sync.RWMutex{},
	}
}

func (c *InMemoryMutex) Add(k interface{}, v interface{}, expiring time.Duration) {
	c.mux.Lock()
	c.items[k] = Item{value: v, timestamp: time.Now(), expiring: expiring}
	c.mux.Unlock()
}

func (c *InMemoryMutex) Get(k interface{}) (interface{}, bool) {
	c.mux.RLock()
	v, ok := c.items[k]
	c.mux.RUnlock()

	// not exists or expired, remove it and return nil if expired
	if !ok || v.expiring != NEVER_EXPIRED && time.Now().After(v.timestamp.Add(v.expiring)) {
		c.Delete(k)
		return nil, false
	}

	return v.value, true
}

func (c *InMemoryMutex) Delete(k interface{}) {
	c.mux.Lock()
	delete(c.items, k)
	c.mux.Unlock()
	return
}
