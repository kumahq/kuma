package mesh

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	cache              *once.Cache
	rm                 manager.ReadOnlyResourceManager
	types              []core_model.ResourceType
	ipFunc             lookup.LookupIPFunc
	meshContextBuilder MeshContextBuilder

	latestMeshContext *xds_context.MeshContext
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	expirationTime time.Duration,
	types []core_model.ResourceType,
	ipFunc lookup.LookupIPFunc,
	meshContextBuilder MeshContextBuilder,
	metrics metrics.Metrics,
) (*Cache, error) {
	c, err := once.New(expirationTime, "mesh_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &Cache{
		rm:                 rm,
		types:              types,
		ipFunc:             ipFunc,
		cache:              c,
		meshContextBuilder: meshContextBuilder,
	}, nil
}

func (c *Cache) GetMeshContext(ctx context.Context, syncLog logr.Logger, mesh string) (xds_context.MeshContext, error) {
	elt, err := c.cache.GetOrRetrieve(ctx, mesh, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		snapshot, err := BuildMeshSnapshot(ctx, key, c.rm, c.types, c.ipFunc)
		if err != nil {
			return nil, err
		}
		snapshotHash := snapshot.Hash()
		if c.latestMeshContext != nil && c.latestMeshContext.Hash == snapshotHash {
			return *c.latestMeshContext, nil
		}

		meshCtx, err := c.meshContextBuilder.Build(snapshot)
		if err != nil {
			return nil, err
		}
		c.latestMeshContext = &meshCtx
		return meshCtx, nil
	}))
	if err != nil {
		return xds_context.MeshContext{}, err
	}
	return elt.(xds_context.MeshContext), nil
}
