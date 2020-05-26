package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func AltStatName() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&altStatNameConfigurer{})
	})
}

type altStatNameConfigurer struct {
}

func (e *altStatNameConfigurer) Configure(cluster *envoy_api.Cluster) error {
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return nil
}
