package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

// Cached version of the ReadOnlyResourceManager designed to be used only for use cases of eventual consistency.
// This cache is NOT consistent across instances of the control plane.
//
// When retrieving elements from cache, they point to the same instances of the resources.
// We cannot do deep copies because it would consume lots of memory, therefore you need to be extra careful to NOT modify the resources.
type cachedManager struct {
	delegate ReadOnlyResourceManager
	cache    *cache.Cache
	metrics  *prometheus.CounterVec

	mutexes  map[string]*sync.Mutex
	mapMutex sync.Mutex // guards "mutexes" field
	tenants  multitenant.Tenants
}

var _ ReadOnlyResourceManager = &cachedManager{}

func NewCachedManager(delegate ReadOnlyResourceManager, expirationTime time.Duration, metrics metrics.Metrics, tenants multitenant.Tenants) (ReadOnlyResourceManager, error) {
	metric := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "store_cache",
		Help: "Summary of Store Cache",
	}, []string{"operation", "resource_type", "result"})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &cachedManager{
		delegate: delegate,
		cache:    cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		metrics:  metric,
		mutexes:  map[string]*sync.Mutex{},
		tenants:  tenants,
	}, nil
}

func (c *cachedManager) Get(ctx context.Context, res model.Resource, fs ...store.GetOptionsFunc) error {
	tenantID, err := c.tenants.GetID(ctx)
	if err != nil {
		return err
	}
	opts := store.NewGetOptions(fs...)
	cacheKey := fmt.Sprintf("GET:%s:%s:%s", res.Descriptor().Name, opts.HashCode(), tenantID)
	obj, found := c.cache.Get(cacheKey)
	if !found {
		// There might be a situation when cache just expired and there are many concurrent goroutines here.
		// We should only let one fill the cache and let the rest of them wait for it. Otherwise we will be repeating expensive work.
		mutex := c.mutexFor(cacheKey)
		mutex.Lock()
		obj, found = c.cache.Get(cacheKey)
		if !found {
			// After many goroutines are unlocked one by one, only one should execute this branch, the rest should retrieve object from the cache
			c.metrics.WithLabelValues("get", string(res.Descriptor().Name), "miss").Inc()
			if err := c.delegate.Get(ctx, res, fs...); err != nil {
				mutex.Unlock()
				return err
			}
			c.cache.SetDefault(cacheKey, res)
		} else {
			c.metrics.WithLabelValues("get", string(res.Descriptor().Name), "hit-wait").Inc()
		}
		mutex.Unlock()
		c.cleanMutexFor(cacheKey) // We need to cleanup mutexes from the map, otherwise we can see the memory leak.
	} else {
		c.metrics.WithLabelValues("get", string(res.Descriptor().Name), "hit").Inc()
	}

	if found {
		cached := obj.(model.Resource)
		if err := res.SetSpec(cached.GetSpec()); err != nil {
			return err
		}
		res.SetMeta(cached.GetMeta())
	}
	return nil
}

func (c *cachedManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	tenantID, err := c.tenants.GetID(ctx)
	if err != nil {
		return err
	}
	opts := store.NewListOptions(fs...)
	if !opts.IsCacheable() {
		return fmt.Errorf("filter functions are not allowed for cached store")
	}
	cacheKey := fmt.Sprintf("LIST:%s:%s:%s", list.GetItemType(), opts.HashCode(), tenantID)
	obj, found := c.cache.Get(cacheKey)
	if !found {
		// There might be a situation when cache just expired and there are many concurrent goroutines here.
		// We should only let one fill the cache and let the rest of them wait for it. Otherwise we will be repeating expensive work.
		mutex := c.mutexFor(cacheKey)
		mutex.Lock()
		obj, found = c.cache.Get(cacheKey)
		if !found {
			// After many goroutines are unlocked one by one, only one should execute this branch, the rest should retrieve object from the cache
			c.metrics.WithLabelValues("list", string(list.GetItemType()), "miss").Inc()
			if err := c.delegate.List(ctx, list, fs...); err != nil {
				mutex.Unlock()
				return err
			}
			c.cache.SetDefault(cacheKey, list.GetItems())
		} else {
			c.metrics.WithLabelValues("list", string(list.GetItemType()), "hit-wait").Inc()
		}
		mutex.Unlock()
		c.cleanMutexFor(cacheKey) // We need to cleanup mutexes from the map, otherwise we can see the memory leak.
	} else {
		c.metrics.WithLabelValues("list", string(list.GetItemType()), "hit").Inc()
	}

	if found {
		resources := obj.([]model.Resource)
		for _, res := range resources {
			if err := list.AddItem(res); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *cachedManager) mutexFor(key string) *sync.Mutex {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	mutex, exist := c.mutexes[key]
	if !exist {
		mutex = &sync.Mutex{}
		c.mutexes[key] = mutex
	}
	return mutex
}

func (c *cachedManager) cleanMutexFor(key string) {
	c.mapMutex.Lock()
	delete(c.mutexes, key)
	c.mapMutex.Unlock()
}
