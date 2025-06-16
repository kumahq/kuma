package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type AltStatNameConfigurer struct{
	StatName string
}

var _ ClusterConfigurer = &AltStatNameConfigurer{}

func (e *AltStatNameConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	if e.StatName != "" {
		cluster.AltStatName = e.StatName
		return nil
	}
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return nil
}
