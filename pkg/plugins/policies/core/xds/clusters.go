package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tags"
	gateway_metadata "github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/metadata"
	generator_metadata "github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

type Clusters struct {
	Inbound       map[string]*envoy_cluster.Cluster
	Outbound      map[string]*envoy_cluster.Cluster
	OutboundSplit map[string][]*envoy_cluster.Cluster
	Gateway       map[string]*envoy_cluster.Cluster
	Egress        map[string]*envoy_cluster.Cluster
	Prometheus    *envoy_cluster.Cluster
}

func GatherAllClusters(rs *core_xds.ResourceSet) []*envoy_cluster.Cluster {
	var clusters []*envoy_cluster.Cluster
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		cluster := res.Resource.(*envoy_cluster.Cluster)
		clusters = append(clusters, cluster)
	}
	return clusters
}

func GatherClusters(rs *core_xds.ResourceSet) Clusters {
	clusters := Clusters{
		Inbound:       map[string]*envoy_cluster.Cluster{},
		Outbound:      map[string]*envoy_cluster.Cluster{},
		OutboundSplit: map[string][]*envoy_cluster.Cluster{},
		Gateway:       map[string]*envoy_cluster.Cluster{},
		Egress:        map[string]*envoy_cluster.Cluster{},
	}
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		cluster := res.Resource.(*envoy_cluster.Cluster)

		switch res.Origin {
		case generator_metadata.OriginOutbound:
			serviceName := tags.ServiceFromClusterName(cluster.Name)
			if serviceName != cluster.Name {
				// first group is service name and second split number
				clusters.OutboundSplit[serviceName] = append(clusters.OutboundSplit[serviceName], cluster)
			} else {
				clusters.Outbound[cluster.Name] = cluster
			}
		case generator_metadata.OriginInbound:
			clusters.Inbound[cluster.Name] = cluster
		case gateway_metadata.OriginGateway:
			clusters.Gateway[cluster.Name] = cluster
		case generator_metadata.OriginEgress:
			clusters.Egress[cluster.Name] = cluster
		case generator_metadata.OriginPrometheus:
			clusters.Prometheus = cluster
		default:
			continue
		}
	}
	return clusters
}
