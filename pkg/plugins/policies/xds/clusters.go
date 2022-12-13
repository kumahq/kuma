package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	clusters_builder "github.com/kumahq/kuma/pkg/xds/envoy/clusters"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

type Clusters struct {
	Inbound  map[string]*envoy_cluster.Cluster
	Outbound map[string]*envoy_cluster.Cluster
	Gateway  map[string]*envoy_cluster.Cluster
}

func GatherClusters(rs *xds.ResourceSet) Clusters {
	clusters := Clusters{
		Inbound:  map[string]*envoy_cluster.Cluster{},
		Outbound: map[string]*envoy_cluster.Cluster{},
		Gateway:  map[string]*envoy_cluster.Cluster{},
	}
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		cluster := res.Resource.(*envoy_cluster.Cluster)

		switch res.Origin {
		case generator.OriginOutbound:
			clusters.Outbound[cluster.Name] = cluster
		case generator.OriginInbound:
			clusters.Inbound[cluster.Name] = cluster
		case metadata.OriginGateway:
			clusters.Gateway[cluster.Name] = cluster
		default:
			continue
		}
	}
	return clusters
}

type NameConfigurer struct {
	Name string
}

func (n *NameConfigurer) Configure(c *envoy_cluster.Cluster) error {
	c.Name = n.Name
	return nil
}

func WithName(name string) clusters_builder.ClusterBuilderOpt {
	return clusters_builder.ClusterBuilderOptFunc(func(config *clusters_builder.ClusterBuilderConfig) {
		config.AddV3(&NameConfigurer{Name: name})
	})
}
