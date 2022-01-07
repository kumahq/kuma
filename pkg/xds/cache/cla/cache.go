package cla

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
)

// Cache is needed to share and cache ClusterLoadAssignments among goroutines
// which reconcile Dataplane's state. In scope of one mesh ClusterLoadAssignment
// will be the same for each service so no need to reconcile for each dataplane.
type Cache struct {
	cache *once.Cache
}

func NewCache(
	expirationTime time.Duration,
	metrics metrics.Metrics,
) (*Cache, error) {
	c, err := once.New(expirationTime, "cla_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &Cache{
		cache: c,
	}, nil
}

func (c *Cache) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion envoy_common.APIVersion, endpointMap xds.EndpointMap) (proto.Message, error) {
	key := sha256.Hash(fmt.Sprintf("%s:%s:%s:%s", apiVersion, meshName, cluster.Hash(), meshHash))

	elt, err := c.cache.GetOrRetrieve(ctx, key, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		matchTags := map[string]string{}
		for tag, val := range cluster.Tags() {
			if tag != mesh_proto.ServiceTag {
				matchTags[tag] = val
			}
		}

		// For the majority of cases we don't have custom tags, we can just take a slice
		endpoints := endpointMap[cluster.Service()]
		if len(matchTags) > 0 {
			endpoints = []xds.Endpoint{}
			for _, endpoint := range endpointMap[cluster.Service()] {
				if endpoint.ContainsTags(matchTags) {
					endpoints = append(endpoints, endpoint)
				}
			}
		}
		return envoy_endpoints.CreateClusterLoadAssignment(cluster.Name(), endpoints, apiVersion)
	}))
	if err != nil {
		return nil, err
	}
	return elt.(proto.Message), nil
}
