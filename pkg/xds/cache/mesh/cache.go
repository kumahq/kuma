package mesh

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	cache  *once.Cache
	rm     manager.ReadOnlyResourceManager
	types  []core_model.ResourceType
	ipFunc lookup.LookupIPFunc
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	expirationTime time.Duration,
	types []core_model.ResourceType,
	ipFunc lookup.LookupIPFunc,
	metrics metrics.Metrics,
) (*Cache, error) {
	c, err := once.New(expirationTime, "mesh_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &Cache{
		rm:     rm,
		types:  types,
		ipFunc: ipFunc,
		cache:  c,
	}, nil
}

func (c *Cache) GetHash(ctx context.Context, mesh string) (string, error) {
	elt, err := c.cache.GetOrRetrieve(ctx, mesh, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		snapshot, err := GetMeshSnapshot(ctx, key, c.rm, c.types, c.ipFunc)
		if err != nil {
			return nil, err
		}
		return snapshot.hash(), nil
	}))
	if err != nil {
		return "", err
	}
	return elt.(string), nil
}
