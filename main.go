package cache

import (
	"container/list"
	"sync"
	"time"
)

type BurnStrategy int

const (
	BurnStrategyRandom    = BurnStrategy(1)
	BurnStrategyOldest    = BurnStrategy(2)
	BurnStrategyOldestLRU = BurnStrategy(3)
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
	BurnStrategy   BurnStrategy
	ExpirationTime time.Duration
}

func NewCache(c CacheOptions) *Cache {
	newc := &Cache{}
	newc.options = c
	newc.contents = make(map[string]*list.Element)
	newc.mutex = &sync.Mutex{}
	newc.l = list.New()
	if newc.options.MaxEntries > newc.options.Upper && newc.options.Upper > 0 {
		newc.options.Upper = newc.options.MaxEntries
	}
	return newc
}

func (c *Cache) Set(key string, value interface{}) {
	c.lock()
	defer c.unlock()
	if value == nil {
		c.deleteItem(c.contents[key])
		return
	}
	if c.l.Len()+1 > c.options.Upper && c.options.Upper > 0 {
		c.burnEntryByStrategy()
	}
	newitem := &CachedItem{}
	newitem.value = value
	newitem.key = key
	e := c.l.PushFront(newitem)
	c.contents[key] = e
	if c.options.ExpirationTime > 0 {
		go c.expireIn(c.options.ExpirationTime, e)
	}
}

func (c *Cache) Get(key string) interface{} {
	c.lock()
	defer c.unlock()
	k, ok := c.contents[key]
	if ok {
		c.hits++
		if c.options.BurnStrategy == BurnStrategyOldestLRU {
			c.l.MoveToFront(k)
		}
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

func (c *Cache) RemoveItem(key string) {
	c.lock()
	defer c.unlock()
	c.deleteItem(c.contents[key])
}

func (c *Cache) Trim(num int) {
	c.lock()
	defer c.unlock()
	for num > 0 && c.l.Len() > 0 {
		c.burnEntryByStrategy()
	}
}

func (c *Cache) Bump(key string) {
	k, ok := c.contents[key]
	if ok {
		c.l.MoveToFront(k)
	}
}

//For now we, can't use this internally since it is mutex'd
func (c *Cache) Len() int {
	c.lock()
	defer c.unlock()
	return c.l.Len()
}

//private functions
func (c *Cache) burnEntryByStrategy() {
	if c.options.BurnStrategy == BurnStrategyOldest || c.options.BurnStrategy == BurnStrategyOldestLRU {
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
	c.l.Remove(i)
	delete(c.contents, i.Value.(*CachedItem).key)
}
