package clusters

import (
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func AltStatName() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(config *ClusterBuilderConfig) {
		config.Add(&AltStatNameConfigurer{
		})
	})
}

type AltStatNameConfigurer struct {
	Name string
}

func (e *AltStatNameConfigurer) Configure(cluster *v2.Cluster) error {
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return nil
}
