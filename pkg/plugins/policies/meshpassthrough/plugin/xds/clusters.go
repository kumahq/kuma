package xds

import (
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

func CreateCluster(apiVersion core_xds.APIVersion, name string, protocol core_meta.Protocol) (envoy.NamedResource, error) {
	clusterBuilder := xds_clusters.NewClusterBuilder(apiVersion, name).
		Configure(xds_clusters.PassThroughCluster()).
		Configure(xds_clusters.DefaultTimeout())
	switch protocol {
	case core_meta.ProtocolGRPC, core_meta.ProtocolHTTP2:
		clusterBuilder.Configure(xds_clusters.Http2())
	}
	return clusterBuilder.Build()
}
