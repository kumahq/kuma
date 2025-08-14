package xds

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/origin"
	gateway_metadata "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	generator_metadata "github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

type EndpointMap map[xds.ServiceName][]*endpointv3.ClusterLoadAssignment

func GatherOutboundEndpoints(rs *xds.ResourceSet) EndpointMap {
	return gatherEndpoints(rs, generator_metadata.OriginOutbound)
}

func GatherGatewayEndpoints(rs *xds.ResourceSet) EndpointMap {
	return gatherEndpoints(rs, gateway_metadata.OriginGateway)
}

func GatherEgressEndpoints(rs *xds.ResourceSet) EndpointMap {
	return gatherEndpoints(rs, generator_metadata.OriginEgress)
}

func gatherEndpoints(rs *xds.ResourceSet, origin origin.Origin) EndpointMap {
	em := EndpointMap{}
	for _, res := range rs.Resources(envoy_resource.EndpointType) {
		if res.Origin != origin {
			continue
		}

		cla := res.Resource.(*endpointv3.ClusterLoadAssignment)
		serviceName := tags.ServiceFromClusterName(cla.ClusterName)
		em[serviceName] = append(em[serviceName], cla)
	}
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		if res.Origin != origin {
			continue
		}

		cluster := res.Resource.(*clusterv3.Cluster)
		serviceName := tags.ServiceFromClusterName(cluster.Name)
		if cluster.LoadAssignment != nil {
			em[serviceName] = append(em[serviceName], cluster.LoadAssignment)
		}
	}
	return em
}
