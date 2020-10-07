package mesh

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/xds/cache/once"

	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var (
	meshCacheLog = core.Log.WithName("mesh-cache")
)

type Cache struct {
	cache   *cache.Cache
	rm      manager.ReadOnlyResourceManager
	types   []core_model.ResourceType
	ipFunc  lookup.LookupIPFunc
	onceMap *once.Map
}

func NewCache(rm manager.ReadOnlyResourceManager, expirationTime time.Duration, types []core_model.ResourceType, ipFunc lookup.LookupIPFunc) *Cache {
	return &Cache{
		rm:      rm,
		types:   types,
		ipFunc:  ipFunc,
		onceMap: once.NewMap(),
		cache:   cache.New(expirationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}
}

func (c *Cache) GetHash(ctx context.Context, mesh string) (string, error) {
	hash, found := c.cache.Get(mesh)
	if found {
		return hash.(string), nil
	}
	o := c.onceMap.Get(mesh)
	o.Do(func() (interface{}, error) {
		snapshot, err := GetMeshSnapshot(ctx, mesh, c.rm, c.types, c.ipFunc)
		if err != nil {
			return nil, err
		}
		hash = snapshot.hash()
		c.cache.SetDefault(mesh, hash)
		c.onceMap.Delete(mesh)
		return hash, nil
	})
	if o.Err != nil {
		return "", o.Err
	}
	return o.Value.(string), nil
}
