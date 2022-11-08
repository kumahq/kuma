package mesh

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/patrickmn/go-cache"

	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	// cache is used for caching a context and ignoring mesh changes for up to a
	// short expiration time.
	cache *once.Cache
	// hashCache keeps a cached context, for a much longer time, that is only reused
	// when the mesh hasn't changed.
	hashCache *cache.Cache

	meshContextBuilder xds_context.MeshContextBuilder
}

// recalculationTime is the time after which the mesh context is recalculated no
// matter what.
// It exists to ensure contexts of deleted Meshes are eventually cleaned up.
const recalculationTime = time.Hour

func NewCache(
	expirationTime time.Duration,
	meshContextBuilder xds_context.MeshContextBuilder,
	metrics metrics.Metrics,
) (*Cache, error) {
	c, err := once.New(expirationTime, "mesh_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &Cache{
		cache:              c,
		meshContextBuilder: meshContextBuilder,
		hashCache:          cache.New(recalculationTime, time.Duration(int64(float64(expirationTime)*0.9))),
	}, nil
}

func (c *Cache) GetMeshContext(ctx context.Context, syncLog logr.Logger, mesh string) (xds_context.MeshContext, error) {
	elt, err := c.cache.GetOrRetrieve(ctx, mesh, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		// Check hashCache first for an existing mesh context
		var context xds_context.MeshContext
		if cached, ok := c.hashCache.Get(mesh); ok {
			context = *cached.(*xds_context.MeshContext)
		} else {
			// If we don't have any context, we build one, set it in
			// `hashCache` and return it.
			meshCtx, err := c.meshContextBuilder.Build(ctx, mesh)
			if err != nil {
				return xds_context.MeshContext{}, err
			}
			c.hashCache.Set(mesh, &meshCtx, 0)
			return meshCtx, nil
		}

		// If we have some context, use it only if the hash hasn't changed.
		// Otherwise, rebuild it.
		meshCtx, err := c.meshContextBuilder.BuildIfChanged(ctx, mesh, context.Hash)
		if err != nil {
			return xds_context.MeshContext{}, err
		}
		if meshCtx == nil {
			// Context didn't need to be rebuilt
			return context, nil
		}

		c.hashCache.Set(mesh, meshCtx, 0)
		return *meshCtx, nil
	}))
	if err != nil {
		return xds_context.MeshContext{}, err
	}
	return elt.(xds_context.MeshContext), nil
}
