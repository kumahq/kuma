package generator

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator/routes"
)

type destinations map[string]map[string][]envoy_tags.Tags

func (d destinations) get(mesh string, service string) []envoy_tags.Tags {
	forMesh := d[mesh]
	return append(forMesh[service], forMesh[mesh_proto.MatchAllTag]...)
}

func buildDestinations(ingressProxy *core_xds.ZoneIngressProxy) destinations {
	dest := destinations{}
	availableSvcsByMesh := map[string][]*mesh_proto.ZoneIngress_AvailableService{}
	for _, service := range ingressProxy.ZoneIngressResource.Spec.AvailableServices {
		availableSvcsByMesh[service.Mesh] = append(availableSvcsByMesh[service.Mesh], service)
	}
	for _, meshResources := range ingressProxy.MeshResourceList {
		res := xds_context.Resources{MeshLocalResources: meshResources.Resources}
		destForMesh := map[string][]envoy_tags.Tags{}
		meshHTTPRoutes := res.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).(*meshhttproute_api.MeshHTTPRouteResourceList).Items
		addTrafficRouteDestinations(res.TrafficRoutes().Items, destForMesh)
		routes.AddMeshHTTPRouteDestinations(meshHTTPRoutes, destForMesh)
		addGatewayRouteDestinations(res.GatewayRoutes().Items, destForMesh)
		addMeshGatewayDestinations(res.MeshGateways().Items, destForMesh)
		addVirtualOutboundDestinations(res.VirtualOutbounds().Items, availableSvcsByMesh[meshResources.Mesh.GetMeta().GetName()], destForMesh)
		dest[meshResources.Mesh.GetMeta().GetName()] = destForMesh
	}
	return dest
}

func addMeshGatewayDestinations(
	meshGateways []*core_mesh.MeshGatewayResource,
	destinations map[string][]envoy_tags.Tags,
) {
	for _, meshGateway := range meshGateways {
		for _, selector := range meshGateway.Selectors() {
			addMeshGatewayListenersDestinations(
				meshGateway.Spec,
				selector.GetMatch(),
				destinations,
			)
		}
	}
}

func addMeshGatewayListenersDestinations(
	meshGateway *mesh_proto.MeshGateway,
	matchTags map[string]string,
	destinations map[string][]envoy_tags.Tags,
) {
	service := matchTags[mesh_proto.ServiceTag]

	for _, listener := range meshGateway.GetConf().GetListeners() {
		if !listener.CrossMesh {
			continue
		}

		destinations[service] = append(
			destinations[service],
			mesh_proto.Merge(
				meshGateway.GetTags(),
				matchTags,
				listener.GetTags(),
			),
		)
	}
}

func addGatewayRouteDestinations(
	gatewayRoutes []*core_mesh.MeshGatewayRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	var backends []*mesh_proto.MeshGatewayRoute_Backend

	for _, route := range gatewayRoutes {
		for _, rule := range route.Spec.GetConf().GetHttp().GetRules() {
			backends = append(backends, rule.Backends...)
		}

		for _, rule := range route.Spec.GetConf().GetTcp().GetRules() {
			backends = append(backends, rule.Backends...)
		}
	}

	for _, backend := range backends {
		routes.AddDestination(backend.Destination, destinations)
	}
}

func addTrafficRouteDestinations(
	policies []*core_mesh.TrafficRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	for _, policy := range policies {
		for _, split := range policy.Spec.Conf.GetSplitWithDestination() {
			routes.AddDestination(split.Destination, destinations)
		}

		for _, http := range policy.Spec.Conf.Http {
			for _, split := range http.GetSplitWithDestination() {
				routes.AddDestination(split.Destination, destinations)
			}
		}
	}
}

func addVirtualOutboundDestinations(
	virtualOutbounds []*core_mesh.VirtualOutboundResource,
	availableServices []*mesh_proto.ZoneIngress_AvailableService,
	destinations map[string][]envoy_tags.Tags,
) {
	// If there are no VirtualOutbounds, we are not modifying destinations
	if len(virtualOutbounds) == 0 {
		return
	}

	for _, availableService := range availableServices {
		for _, matched := range dns.Match(virtualOutbounds, availableService.Tags) {
			service := availableService.Tags[mesh_proto.ServiceTag]
			tags := envoy_tags.Tags{}
			for _, param := range matched.Spec.GetConf().GetParameters() {
				tags[param.TagKey] = availableService.Tags[param.TagKey]
			}
			destinations[service] = append(destinations[service], tags)
		}
	}
}
