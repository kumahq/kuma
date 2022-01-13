package mesh

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	cache              *once.Cache
	meshContextBuilder xds_context.MeshContextBuilder

	latestMeshContext *xds_context.MeshContext
}

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
	}, nil
}

func (c *Cache) GetMeshContext(ctx context.Context, syncLog logr.Logger, mesh string) (xds_context.MeshContext, error) {
	elt, err := c.cache.GetOrRetrieve(ctx, mesh, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		if c.latestMeshContext == nil {
			meshCtx, err := c.meshContextBuilder.Build(ctx, mesh)
			if err != nil {
				return xds_context.MeshContext{}, err
			}
			c.latestMeshContext = &meshCtx
		} else {
			meshCtx, err := c.meshContextBuilder.BuildIfChanged(ctx, mesh, c.latestMeshContext.Hash)
			if err != nil {
				return xds_context.MeshContext{}, err
			}
			if meshCtx != nil {
				c.latestMeshContext = meshCtx
			}
		}
		return *c.latestMeshContext, nil
	}))
	if err != nil {
		return xds_context.MeshContext{}, err
	}
	return elt.(xds_context.MeshContext), nil
}
