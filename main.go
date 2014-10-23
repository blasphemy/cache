package cache

import (
	"sync"
	"time"
)

const (
	BurnStrategyRandom = 1
	BurnStrategyOldest = 2
)

type Cache struct {
	contents map[string]*CachedItem
	mutex    *sync.Mutex
	dead     bool
	hits     int64
	misses   int64
	options  CacheOptions
}

type CachedItem struct {
	ts    time.Time
	value interface{}
}

type CacheOptions struct {
	MaxEntries     int //If this is set to 0, you must have an expiration time set.
	Upper          int
	BurnStrategy   int
	ExpirationTime time.Duration
}

func NewCache(c CacheOptions) *Cache {
	newc := &Cache{}
	newc.options = c
	newc.contents = make(map[string]*CachedItem)
	newc.mutex = &sync.Mutex{}
	return newc
}

func (c *Cache) Set(key string, value interface{}) {
	c.lock()
	defer c.unlock()
	for c.options.Upper > 0 && len(c.contents) > c.options.Upper {
		c.burnEntryByStrategy()
	}
	newitem := &CachedItem{}
	newitem.ts = time.Now()
	newitem.value = value
	c.contents[key] = newitem
}

func (c *Cache) Get(key string) interface{} {
	c.lock()
	defer c.unlock()
	k := c.contents[key]
	if k != nil {
		c.hits++
		return k.value
	} else {
		c.misses++
		return nil
	}
}

func (c *Cache) Start() {
	go c.runner()
}

func (c *Cache) Stop() {
	c.dead = true
}

func (c *Cache) Hits() int64 {
	return c.hits
}

func (c *Cache) Misses() int64 {
	return c.misses
}

//private functions
func (c *Cache) burnEntryByStrategy() {
	if c.options.BurnStrategy == BurnStrategyOldest {
		c.burnEntryByOldest()
	} else {
		c.burnEntryByRandom()
	}
}

func (c *Cache) burnEntryByRandom() {
	for a, _ := range c.contents {
		delete(c.contents, a)
		break
	}
}

func (c *Cache) burnEntryByOldest() {
	var ts time.Time
	var i string
	//seed ts from random item
	for a, b := range c.contents {
		i = a
		ts = b.ts
		break
	}
	for a, b := range c.contents {
		if b.ts < ts {
			i = a
			ts = b.ts
		}
	}
}

func (c *Cache) burnExpiredKeys() {
	for a, k := range c.contents {
		if time.Since(k.ts) > c.options.ExpirationTime && c.options.ExpirationTime > 0 {
			delete(c.contents, a)
		}
	}
}

func (c *Cache) lock() {
	c.mutex.Lock()
}

func (c *Cache) unlock() {
	c.mutex.Unlock()
}

func (c *Cache) runner() {
	for {
		if c.dead {
			break
		}
		c.lock()
		if c.options.MaxEntries > 0 {
			for len(c.contents) > c.options.MaxEntries {
				c.burnEntryByStrategy()
			}
		}
		c.unlock()
	}
}
