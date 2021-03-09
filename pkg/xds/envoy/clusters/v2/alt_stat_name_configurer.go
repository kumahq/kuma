package clusters

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type AltStatNameConfigurer struct {
}

var _ ClusterConfigurer = &AltStatNameConfigurer{}

func (e *AltStatNameConfigurer) Configure(cluster *envoy_api.Cluster) error {
	sanitizedName := util_xds.SanitizeMetric(cluster.Name)
	if sanitizedName != cluster.Name {
		cluster.AltStatName = sanitizedName
	}
	return nil
}
