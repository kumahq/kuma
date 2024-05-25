package xds

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

func CreateCluster(apiVersion core_xds.APIVersion, name string) (envoy.NamedResource, error) {
	return envoy_clusters.NewClusterBuilder(apiVersion, name).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
}
