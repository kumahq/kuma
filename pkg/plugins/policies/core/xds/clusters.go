package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

type Clusters struct {
	Inbound       map[string]*envoy_cluster.Cluster
	Outbound      map[string]*envoy_cluster.Cluster
	OutboundSplit map[string][]*envoy_cluster.Cluster
	Gateway       map[string]*envoy_cluster.Cluster
	Egress        map[string]*envoy_cluster.Cluster
	Prometheus    *envoy_cluster.Cluster
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
		case generator.OriginOutbound:
			serviceName := tags.ServiceFromClusterName(cluster.Name)
			if serviceName != cluster.Name {
				// first group is service name and second split number
				clusters.OutboundSplit[serviceName] = append(clusters.OutboundSplit[serviceName], cluster)
			} else {
				clusters.Outbound[cluster.Name] = cluster
			}
		case generator.OriginInbound:
			clusters.Inbound[cluster.Name] = cluster
		case metadata.OriginGateway:
			clusters.Gateway[cluster.Name] = cluster
		case egress.OriginEgress:
			clusters.Egress[cluster.Name] = cluster
		case generator.OriginPrometheus:
			clusters.Prometheus = cluster
		default:
			continue
		}
	}
	return clusters
}

func GatherTargetedClusters(
	outbounds []*core_xds.Outbound,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	outboundClusters map[string]*envoy_cluster.Cluster,
) map[*envoy_cluster.Cluster]string {
	targetedClusters := map[*envoy_cluster.Cluster]string{}
	for _, outbound := range outbounds {
		serviceName := outbound.LegacyOutbound.GetService()
		for _, splitCluster := range outboundSplitClusters[serviceName] {
			targetedClusters[splitCluster] = serviceName
		}

		cluster, ok := outboundClusters[serviceName]
		if ok {
			targetedClusters[cluster] = serviceName
		}
	}

	return targetedClusters
}
