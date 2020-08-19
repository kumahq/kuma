package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

// Cached version of the ReadOnlyResourceManager designed to be used only for use cases of eventual consistency.
//
// This cache is NOT consistent across instances of the control plane.
// This cache is mutex free for performance with consideration that: values can be overridden by other goroutine.
// * if cache expires and multiple goroutines tries to fetch the resource, they may fetch underlying manager multiple times.
// * if cache expires and multiple goroutines tries to fetch the resource, they fetch underlying manager multiple times
//   and the value returned is different, the older value may be persisted. This is ok, since this cache is designed
//   to have low expiration time (like 1s) and having old value just extends propagation of new config for 1 more second.
type cachedManager struct {
	delegate ReadOnlyResourceManager
	cache    *cache.Cache
	metrics  *prometheus.CounterVec
}

var _ ReadOnlyResourceManager = &cachedManager{}

func NewCachedManager(delegate ReadOnlyResourceManager, expirationTime time.Duration) ReadOnlyResourceManager {
	return &cachedManager{
		delegate: delegate,
		cache:    cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
		metrics: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "store_cache",
			Help: "Summary of Store Cache",
		}, []string{"operation", "resource_type", "result"}),
	}
}

func (c cachedManager) Get(ctx context.Context, res model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)
	cacheKey := fmt.Sprintf("GET:%s:%s", res.GetType(), opts.HashCode())
	obj, found := c.cache.Get(cacheKey)
	if !found {
		c.metrics.WithLabelValues("get", string(res.GetType()), "miss").Inc()
		if err := c.delegate.Get(ctx, res, fs...); err != nil {
			return err
		}
		c.cache.SetDefault(cacheKey, res)
	} else {
		c.metrics.WithLabelValues("get", string(res.GetType()), "hit").Inc()
		cached := obj.(model.Resource)
		if err := res.SetSpec(cached.GetSpec()); err != nil {
			return err
		}
		res.SetMeta(cached.GetMeta())
	}
	return nil
}

func (c cachedManager) List(ctx context.Context, list model.ResourceList, fs ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(fs...)
	cacheKey := fmt.Sprintf("LIST:%s:%s", list.GetItemType(), opts.HashCode())
	obj, found := c.cache.Get(cacheKey)
	if !found {
		c.metrics.WithLabelValues("list", string(list.GetItemType()), "miss").Inc()
		if err := c.delegate.List(ctx, list, fs...); err != nil {
			return err
		}
		c.cache.SetDefault(cacheKey, list.GetItems())
	} else {
		c.metrics.WithLabelValues("list", string(list.GetItemType()), "hit").Inc()
		resources := obj.([]model.Resource)
		for _, res := range resources {
			if err := list.AddItem(res); err != nil {
				return err
			}
		}
	}
	return nil
}
