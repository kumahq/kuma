package xds

import (
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

type EndpointMap map[xds.ServiceName][]*endpointv3.ClusterLoadAssignment

func GatherEndpoints(rs *xds.ResourceSet) EndpointMap {
	em := EndpointMap{}
	for _, res := range rs.Resources(envoy_resource.EndpointType) {
		if res.Origin != generator.OriginOutbound {
			continue
		}

		cla := res.Resource.(*endpointv3.ClusterLoadAssignment)
		serviceName := tags.ServiceFromClusterName(cla.ClusterName)
		em[serviceName] = append(em[serviceName], cla)
	}
	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		if res.Origin != generator.OriginOutbound {
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

func GatherEgressEndpoints(rs *xds.ResourceSet) EndpointMap {
	em := EndpointMap{}
	for _, res := range rs.Resources(envoy_resource.EndpointType) {
		if res.Origin != egress.OriginEgress {
			continue
		}

		cla := res.Resource.(*endpointv3.ClusterLoadAssignment)
		serviceName := tags.ServiceFromClusterName(cla.ClusterName)
		em[serviceName] = append(em[serviceName], cla)
	}

	for _, res := range rs.Resources(envoy_resource.ClusterType) {
		if res.Origin != egress.OriginEgress {
			continue
		}

		cluster := res.Resource.(*clusterv3.Cluster)
		if cluster.LoadAssignment != nil {
			em[cluster.Name] = append(em[cluster.Name], cluster.LoadAssignment)
		}
	}
	return em
}
