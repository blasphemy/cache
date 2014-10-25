package cache

import (
	"container/list"
	"sync"
	"time"
)

type CacheStrategy int

const (
	CacheStrategyRandom    = CacheStrategy(1)
	CacheStrategyOldest    = CacheStrategy(2)
	CacheStrategyOldestLRU = CacheStrategy(3)
	CacheStrategyLFU       = CacheStrategy(4)
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
	key     string
	value   interface{}
	running bool
	used    int
}

type CacheOptions struct {
	MaxEntries     int //If this is set to 0, you must have an expiration time set.
	Upper          int
	CacheStrategy  CacheStrategy
	ExpirationTime time.Duration
	JobInvertal    time.Duration
	SafeRange      int
}

func NewCache(c CacheOptions) *Cache {
	newc := &Cache{}
	newc.options = c
	newc.contents = make(map[string]*list.Element)
	newc.mutex = &sync.Mutex{}
	newc.l = list.New()
	newc.dead = true
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
	if c.l.Len()+1 >= c.options.Upper && c.options.Upper > 0 {
		c.burnEntryByStrategy()
	}
	newitem := &CachedItem{}
	newitem.value = value
	newitem.key = key
	e := c.l.PushFront(newitem)
	c.contents[key] = e
	if c.options.ExpirationTime > 0 && !c.dead {
		go c.expireIn(c.options.ExpirationTime, e)
	}
}

func (c *Cache) Get(key string) interface{} {
	c.lock()
	defer c.unlock()
	k, ok := c.contents[key]
	if ok {
		c.hits++
		if c.options.CacheStrategy == CacheStrategyOldestLRU || c.options.CacheStrategy == CacheStrategyLFU {
			c.l.MoveToFront(k)
			k.Value.(*CachedItem).used++
		}
		return k.Value.(*CachedItem).value
	} else {
		c.misses++
		return nil
	}
}

func (c *Cache) Start() {
	c.dead = false
	if c.options.JobInvertal > 0 {
		go c.runner()
	}
	for _, b := range c.contents {
		if !b.Value.(*CachedItem).running && c.options.ExpirationTime > 0 {
			go c.expireIn(c.options.ExpirationTime, b)
		}
	}
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
		num--
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
	if c.options.CacheStrategy == CacheStrategyOldest || c.options.CacheStrategy == CacheStrategyOldestLRU {
		c.burnEntryByOldest()
	} else if c.options.CacheStrategy == CacheStrategyLFU {
		c.burnEntryByLFU()
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

func (c *Cache) burnEntryByLFU() {
	counter := 0
	for e := c.l.Back(); e != nil; e = e.Prev() {
		if e.Value.(*CachedItem).used == counter {
			c.deleteItem(e)
			break
		}
		counter++
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
		time.Sleep(c.options.JobInvertal)
		if c.dead {
			break
		}
		c.lock()
		if c.options.MaxEntries > 0 {
			for len(c.contents) > c.options.MaxEntries-c.options.SafeRange {
				c.burnEntryByStrategy()
			}
		}
		c.unlock()
	}
}

func (c *Cache) expireIn(t time.Duration, i *list.Element) {
	i.Value.(*CachedItem).running = true
	time.Sleep(t)
	if c.dead {
		i.Value.(*CachedItem).running = false
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
