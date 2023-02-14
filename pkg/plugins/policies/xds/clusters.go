package xds

import (
	"regexp"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator"
	envoy_common "github.com/kumahq/kuma/pkg/xds/generator"
)

var (
	splitClusterRegex = regexp.MustCompile("(.*)(-_[0-9+]_$)")
)

type Clusters struct {
	Inbound       map[string]*envoy_cluster.Cluster
	Outbound      map[string]*envoy_cluster.Cluster
	OutboundSplit map[string][]*envoy_cluster.Cluster
	Gateway       map[string]*envoy_cluster.Cluster
}

func GatherClusters(rs *xds.ResourceSet) Clusters {
	clusters := Clusters{
		Inbound:       map[string]*envoy_cluster.Cluster{},
		Outbound:      map[string]*envoy_cluster.Cluster{},
		OutboundSplit: map[string][]*envoy_cluster.Cluster{},
		Gateway:       map[string]*envoy_cluster.Cluster{},
	}
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		cluster := res.Resource.(*envoy_cluster.Cluster)

		switch res.Origin {
		case generator.OriginOutbound:
			matchedGroups := splitClusterRegex.FindStringSubmatch(cluster.Name)
			if len(matchedGroups) == 3 {
				serviceName := matchedGroups[1]
				// first group is service name and second split number
				clusters.OutboundSplit[serviceName] = append(clusters.OutboundSplit[serviceName], cluster)
			} else {
				clusters.Outbound[cluster.Name] = cluster
			}
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

func GatherTargetedClusters(
	outbounds []*mesh_proto.Dataplane_Networking_Outbound,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	outboundClusters map[string]*envoy_cluster.Cluster,
) map[*envoy_cluster.Cluster]string {
	targetedClusters := map[*envoy_cluster.Cluster]string{}
	for _, outbound := range outbounds {
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
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

// InferProtocol infers protocol for the destination listener.
// It will only return HTTP when all endpoints are tagged with HTTP.
func InferProtocol(routing core_xds.Routing, serviceName string) core_mesh.Protocol {
	var allEndpoints []core_xds.Endpoint
	outboundEndpoints := core_xds.EndpointList(routing.OutboundTargets[serviceName])
	allEndpoints = append(allEndpoints, outboundEndpoints...)
	externalEndpoints := routing.ExternalServiceOutboundTargets[serviceName]
	allEndpoints = append(allEndpoints, externalEndpoints...)

	return envoy_common.InferServiceProtocol(allEndpoints)
}

func HasExternalService(routing core_xds.Routing, serviceName string) bool {
	// We assume that all the targets are either ExternalServices or not
	// therefore we check only the first one
	if endpoints := routing.OutboundTargets[serviceName]; len(endpoints) > 0 {
		if endpoints[0].IsExternalService() {
			return true
		}
	}

	if endpoints := routing.ExternalServiceOutboundTargets[serviceName]; len(endpoints) > 0 {
		return true
	}

	return false
}
