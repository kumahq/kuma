package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func CommonLb() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&CommonLbConfigurer{})
		config.Add(&altStatNameConfigurer{})
		config.Add(&timeoutConfigurer{})
	})
}

type CommonLbConfigurer struct {
}

func (clb *CommonLbConfigurer) Configure(c *envoy_api.Cluster) error {
	c.CommonLbConfig = &envoy_api.Cluster_CommonLbConfig{
		LocalityConfigSpecifier: &envoy_api.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
			LocalityWeightedLbConfig: &envoy_api.Cluster_CommonLbConfig_LocalityWeightedLbConfig{},
		},
	}
	return nil
}
