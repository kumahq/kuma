package cla

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/xds"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/metrics"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	claCacheLog = core.Log.WithName("cla-cache")
)

// Cache is needed to share and cache ClusterLoadAssignments among goroutines
// which reconcile Dataplane's state. In scope of one mesh ClusterLoadAssignment
// will be the same for each service so no need to reconcile for each dataplane.
type Cache struct {
	cache  *once.Cache
	rm     manager.ReadOnlyResourceManager
	dsl    datasource.Loader
	ipFunc lookup.LookupIPFunc
	zone   string
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
	dsl datasource.Loader,
	zone string, expirationTime time.Duration,
	ipFunc lookup.LookupIPFunc,
	metrics metrics.Metrics,
) (*Cache, error) {
	c, err := once.New(expirationTime, "cla_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &Cache{
		cache:  c,
		rm:     rm,
		dsl:    dsl,
		zone:   zone,
		ipFunc: ipFunc,
	}, nil
}

func (c *Cache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion envoy_common.APIVersion) (proto.Message, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", apiVersion, meshName, cluster.Name(), meshHash)
	elt, err := c.cache.GetOrRetrieve(ctx, key, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		dataplanes, err := topology.GetDataplanes(claCacheLog, ctx, c.rm, c.ipFunc, meshName)
		if err != nil {
			return nil, err
		}
		mesh := core_mesh.NewMeshResource()
		if err := c.rm.Get(ctx, mesh, core_store.GetByKey(meshName, model.NoMesh)); err != nil {
			return nil, err
		}
		externalServices := &core_mesh.ExternalServiceResourceList{}
		if err := c.rm.List(ctx, externalServices, core_store.ListByMesh(meshName)); err != nil {
			return nil, err
		}
		endpointMap := topology.BuildEndpointMap(mesh, c.zone, dataplanes.Items, externalServices.Items, c.dsl)
		endpoints := []xds.Endpoint{}
		for _, endpoint := range endpointMap[cluster.Service()] {
			add := true
			for cKey, cValue := range cluster.Tags() {
				eValue, ok := endpoint.Tags[cKey]
				if !ok {
					continue
				}
				if cValue != eValue {
					add = false
					break
				}
			}
			if add {
				endpoints = append(endpoints, endpoint)
			}
		}
		return envoy_endpoints.CreateClusterLoadAssignment(cluster.Name(), endpoints, apiVersion)
	}))
	if err != nil {
		return nil, err
	}
	return elt.(proto.Message), nil
}
