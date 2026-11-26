package cache

import (
	"sync"
	"time"

	"avito/pkg"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type CacheDeadlineDelivery struct {
	cache     *expirable.LRU[string, map[int64]struct{}]
	log       pkg.Logger
	evictChan chan map[int64]struct{}
}

func NewLRUDeadlineCache(ttl, size int, log pkg.Logger) *CacheDeadlineDelivery {
	evictChan := make(chan map[int64]struct{}, 10)
	onEvict := func(key string, val map[int64]struct{}) {
		log.Info("Expired batch evicted: key=%s, val=%v", key, val)
		evictChan <- val
	}

	cache := expirable.NewLRU[string, map[int64]struct{}](size, onEvict, time.Duration(ttl)*time.Second)

	var (
		once      sync.Once
		cacheInst *CacheDeadlineDelivery
	)
	once.Do(func() {
		cacheInst = &CacheDeadlineDelivery{
			cache:     cache,
			log:       log,
			evictChan: evictChan,
		}
	})
	return cacheInst
}

func (c *CacheDeadlineDelivery) Add(key string, val map[int64]struct{}) {
	c.cache.Add(key, val)
	c.log.Info("added key: %s, val: %v", key, val)
}

func (c *CacheDeadlineDelivery) Get(key string) (map[int64]struct{}, bool) {
	val, exist := c.cache.Get(key)
	if !exist {
		return nil, false
	}
	return val, true
}

func (c *CacheDeadlineDelivery) Remove(key string) {
	c.cache.Remove(key)
}

func (c *CacheDeadlineDelivery) GetEvictChan() <-chan map[int64]struct{} {
	return c.evictChan
}
