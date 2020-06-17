package clusters

import (
	"strings"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/xds/envoy"
)

// LbSubset is required for MetadataMatch in Weighted Cluster in TCP Proxy to work.
// Subset loadbalancing is used in two use cases
// 1) TrafficRoute for splitting traffic. Example: TrafficRoute that splits 10% of the traffic to version 1 of the service backend and 90% traffic to version 2 of the service backend
// 2) Multiple outbound sections with the same service
//    Example:
//    type: Dataplane
//    networking:
//      outbound:
//      - port: 1234
//        tags:
//          service: backend
//      - port: 1234
//        tags:
//          service: backend
//          version: v1
//    Only one cluster "backend" is generated for such dataplane, but with lb subset by version.
func LbSubset(tags []envoy.Tags) ClusterBuilderOptFunc {
	return func(config *ClusterBuilderConfig) {
		config.Add(&lbSubsetConfigurer{
			tags: tags,
		})
	}
}

type lbSubsetConfigurer struct {
	tags []envoy.Tags
}

func (e *lbSubsetConfigurer) Configure(c *envoy_api.Cluster) error {
	var selectors []*envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector
	subsets := map[string]bool{} // we only need unique subsets
	for _, tags := range e.tags {
		keys := tags.WithoutTag(mesh_proto.ServiceTag).Keys() // service tag is not included in metadata
		if len(keys) == 0 {
			continue
		}
		joinedTags := strings.Join(keys, ",")
		if !subsets[joinedTags] {
			selectors = append(selectors, &envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector{
				Keys: keys,
				// if there is a split by "version", and there is no endpoint with such version we should not fallback to all endpoints of the service
				FallbackPolicy: envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector_NO_FALLBACK,
			})
			subsets[joinedTags] = true
		}
	}
	if len(selectors) > 0 {
		// if lb subset is set, but no label (Kuma's tag) is queried, we should return any endpoint
		c.LbSubsetConfig = &envoy_api.Cluster_LbSubsetConfig{
			FallbackPolicy:  envoy_api.Cluster_LbSubsetConfig_ANY_ENDPOINT,
			SubsetSelectors: selectors,
		}
	}
	return nil
}
