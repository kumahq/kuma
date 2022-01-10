package mesh

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

// Cache is needed to share and cache Hashes among goroutines which
// reconcile Dataplane's state. Calculating hash is a heavy operation
// that requires fetching all the resources belonging to the Mesh.
type Cache struct {
	cache  *once.Cache
	rm     manager.ReadOnlyResourceManager
	types  []core_model.ResourceType
	ipFunc lookup.LookupIPFunc
	zone   string

	latestMeshContext *xds_context.MeshContext
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	expirationTime time.Duration,
	types []core_model.ResourceType,
	ipFunc lookup.LookupIPFunc,
	zone string,
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
		zone:   zone,
	}, nil
}

func (c *Cache) GetMeshContext(ctx context.Context, syncLog logr.Logger, mesh string) (xds_context.MeshContext, error) {
	elt, err := c.cache.GetOrRetrieve(ctx, mesh, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		snapshot, err := GetMeshSnapshot(ctx, key, c.rm, c.types, c.ipFunc)
		if err != nil {
			return nil, err
		}
		snapshotHash := snapshot.hash()
		if c.latestMeshContext != nil && c.latestMeshContext.Hash == snapshotHash {
			return *c.latestMeshContext, nil
		}

		dataplanesList := snapshot.resources[core_mesh.DataplaneType].(*core_mesh.DataplaneResourceList)
		dataplanes := xds_topology.ResolveAddresses(syncLog, c.ipFunc, dataplanesList.Items)

		zoneIngressList := snapshot.resources[core_mesh.ZoneIngressType].(*core_mesh.ZoneIngressResourceList)
		zoneIngresses := xds_topology.ResolveZoneIngressAddresses(syncLog, c.ipFunc, zoneIngressList.Items)

		meshCtx := xds_context.MeshContext{
			Resource:    snapshot.mesh,
			Dataplanes:  dataplanesList,
			Hash:        snapshotHash,
			EndpointMap: xds_topology.BuildEdsEndpointMap(snapshot.mesh, c.zone, dataplanes, zoneIngresses),
		}
		c.latestMeshContext = &meshCtx
		return meshCtx, nil
	}))
	if err != nil {
		return xds_context.MeshContext{}, err
	}
	return elt.(xds_context.MeshContext), nil
}
