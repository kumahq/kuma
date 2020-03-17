package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
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
}

var _ ReadOnlyResourceManager = &cachedManager{}

func NewCachedManager(delegate ReadOnlyResourceManager, expirationTime time.Duration) ReadOnlyResourceManager {
	return &cachedManager{
		delegate: delegate,
		cache:    cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}
}

func (c cachedManager) Get(ctx context.Context, res model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)
	cacheKey := fmt.Sprintf("GET:%s:%s", res.GetType(), opts.HashCode())
	obj, found := c.cache.Get(cacheKey)
	if !found {
		if err := c.delegate.Get(ctx, res, fs...); err != nil {
			return err
		}
		c.cache.SetDefault(cacheKey, res)
	} else {
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
		if err := c.delegate.List(ctx, list, fs...); err != nil {
			return err
		}
		c.cache.SetDefault(cacheKey, list.GetItems())
	} else {
		resources := obj.([]model.Resource)
		for _, res := range resources {
			if err := list.AddItem(res); err != nil {
				return err
			}
		}
	}
	return nil
}
