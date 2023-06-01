package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	clusters_builder "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

type NameConfigurer struct {
	Name string
}

func (n *NameConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = n.Name
	return nil
}

func WithName(name string) clusters_builder.ClusterBuilderOpt {
	return clusters_builder.ClusterBuilderOptFunc(func(builder *clusters_builder.ClusterBuilder) {
		builder.AddConfigurer(&NameConfigurer{Name: name})
	})
}

func ClusterWithName(name string) envoy_common.NamedResource {
	return clusters_builder.NewClusterBuilder(envoy_common.APIV3).
		Configure(WithName(name)).
		MustBuild()
}
