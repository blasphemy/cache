package cache

import (
	"container/list"
	"sync"
	"time"
)

const (
	BurnStrategyRandom = 1
	BurnStrategyOldest = 2
)

type Cache struct {
	l        *list.List
	contents map[string]*list.Element
	mutex    *sync.Mutex
	dead     bool
	hits     int64
	misses   int64
	options  CacheOptions
}

type CachedItem struct {
	key   string
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
	newc.contents = make(map[string]*list.Element)
	newc.mutex = &sync.Mutex{}
	newc.l = list.New()
	return newc
}

func (c *Cache) Set(key string, value interface{}) {
	c.lock()
	defer c.unlock()
	newitem := &CachedItem{}
	newitem.value = value
	newitem.key = key
	e := c.l.PushFront(newitem)
	c.contents[key] = e
	if c.options.ExpirationTime > 0 {
		go c.expireIn(c.options.ExpirationTime, e)
	}
	if c.Len() > c.options.Upper {
		c.burnEntryByStrategy()
	}
}

func (c *Cache) Get(key string) interface{} {
	c.lock()
	defer c.unlock()
	k := c.contents[key]
	if k != nil {
		c.hits++
		return k.Value.(*CachedItem).value
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
	for _, a := range c.contents {
		c.deleteItem(a)
		break
	}
}

func (c *Cache) burnEntryByOldest() {
	i := c.l.Back()
	if i != nil {
		c.deleteItem(i)
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

func (c *Cache) expireIn(t time.Duration, i *list.Element) {
	time.Sleep(t)
	if c.dead {
		return
	}
	c.lock()
	defer c.unlock()
	c.deleteItem(i)
}

func (c *Cache) deleteItem(i *list.Element) {
	c.lock()
	defer c.unlock()
	c.l.Remove(i)
	delete(c.contents, i.Value.(*CachedItem).key)
}

func (c *Cache) Len() int {
	c.lock()
	defer c.unlock()
	return c.l.Len()
}
