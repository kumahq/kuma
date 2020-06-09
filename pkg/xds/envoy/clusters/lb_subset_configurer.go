package clusters

import (
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/xds/envoy"
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func LbSubset(tags []envoy.Tags) ClusterBuilderOptFunc {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&lbSubsetConfigurer{
			tags: tags,
		})
	})
}

type lbSubsetConfigurer struct {
	tags []envoy.Tags
}

func (e *lbSubsetConfigurer) Configure(c *envoy_api.Cluster) error {
	var selectors []*envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector
	for _, tags := range e.tags {
		keys := tags.WithoutTag(v1alpha1.ServiceTag).Keys()
		if len(keys) == 0 {
			continue
		}
		selectors = append(selectors, &envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector{
			Keys:                 keys,
			FallbackPolicy:       envoy_api.Cluster_LbSubsetConfig_LbSubsetSelector_NO_FALLBACK,
		})
	}
	if len(selectors) > 0 {
		c.LbSubsetConfig = &envoy_api.Cluster_LbSubsetConfig{
			FallbackPolicy: envoy_api.Cluster_LbSubsetConfig_ANY_ENDPOINT,
			SubsetSelectors: selectors,
		}
	}
	return nil
}
