package xds

import (
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	clusters_builder "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

func ClusterWithName(name string) envoy_common.NamedResource {
	return clusters_builder.NewClusterBuilder(envoy_common.APIV3, name).MustBuild()
}
