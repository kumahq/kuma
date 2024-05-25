package xds

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

func CreateCluster(apiVersion core_xds.APIVersion, name string) (envoy.NamedResource, error) {
	return xds_clusters.NewClusterBuilder(apiVersion, name).
		Configure(xds_clusters.PassThroughCluster()).
		Configure(xds_clusters.DefaultTimeout()).
		Build()
}
