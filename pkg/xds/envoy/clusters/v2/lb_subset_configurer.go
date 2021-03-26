package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type LbSubsetConfigurer struct {
	TagKeySets envoy.TagKeysSlice
}

var _ ClusterConfigurer = &LbSubsetConfigurer{}

func (e *LbSubsetConfigurer) Configure(c *envoy_api.Cluster) error {
	var selectors []*envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector
	for _, tagSet := range e.TagKeySets {
		selectors = append(selectors, &envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector{
			Keys: tagSet,
			// if there is a split by "version", and there is no endpoint with such version we should not fallback to all endpoints of the service
			FallbackPolicy: envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector_NO_FALLBACK,
		})
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
