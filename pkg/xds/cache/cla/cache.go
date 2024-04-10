package cla

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/xds/cache/once"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
)

// CachedRetriever is needed to share and cache ClusterLoadAssignments among goroutines
// which reconcile Dataplane's state. In scope of one mesh ClusterLoadAssignment
// will be the same for each service so no need to reconcile for each dataplane.
type CachedRetriever struct {
	cache *once.Cache
	r     Retriever
}

func NewCache(
	expirationTime time.Duration,
	metrics metrics.Metrics,
) (*CachedRetriever, error) {
	c, err := once.New(expirationTime, "cla_cache", metrics)
	if err != nil {
		return nil, err
	}
	return &CachedRetriever{
		cache: c,
	}, nil
}

func (c *CachedRetriever) GetCLA(ctx context.Context, meshName, meshHash string, cluster envoy_common.Cluster, apiVersion xds.APIVersion, endpointMap xds.EndpointMap) (proto.Message, error) {
	key := sha256.Hash(fmt.Sprintf("%s:%s:%s:%s", apiVersion, meshName, cluster.Hash(), meshHash))

	elt, err := c.cache.GetOrRetrieve(ctx, key, once.RetrieverFunc(func(ctx context.Context, key string) (interface{}, error) {
		return c.r.GetCLA(ctx, "", "", cluster, apiVersion, endpointMap)
	}))
	if err != nil {
		return nil, err
	}
	return elt.(proto.Message), nil
}

type Retriever struct{}

func (r *Retriever) GetCLA(_ context.Context, _, _ string, cluster envoy_common.Cluster, apiVersion xds.APIVersion, endpointMap xds.EndpointMap) (proto.Message, error) {
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
}
