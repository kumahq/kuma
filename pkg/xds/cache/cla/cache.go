package cla

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

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
	ipFunc lookup.LookupIPFunc
	zone   string
}

func NewCache(
	rm manager.ReadOnlyResourceManager,
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
		zone:   zone,
		ipFunc: ipFunc,
	}, nil
}

func (c *Cache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion envoy_common.APIVersion) (proto.Message, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", apiVersion, meshName, cluster.Hash(), meshHash)
	elt, err := c.cache.GetOrRetrieve(ctx, key, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		dataplanes, err := topology.GetDataplanes(claCacheLog, ctx, c.rm, c.ipFunc, meshName)
		if err != nil {
			return nil, err
		}
		mesh := core_mesh.NewMeshResource()
		if err := c.rm.Get(ctx, mesh, core_store.GetByKey(meshName, model.NoMesh)); err != nil {
			return nil, err
		}
		// We pick here EndpointMap without External Services
		//
		// This also solves the problem that if the ExternalService is blocked by TrafficPermission
		// OutboundProxyGenerate treats this as EDS cluster and tries to get endpoints via GetCLA
		// Since GetCLA is consistent for a mesh, it would return an endpoint with address which is not valid for EDS.
		endpointMap := topology.BuildEdsEndpointMap(mesh, c.zone, dataplanes.Items)
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
