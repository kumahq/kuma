package zoneproxy

import (
	"reflect"

	"golang.org/x/exp/slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/dns"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type MeshDestinations struct {
	KumaIoServices map[string][]envoy_tags.Tags
	BackendRefs    []BackendRefDestination
}

type BackendRefDestination struct {
	Mesh string
	// DestinationName is a string to reference Type+Name+Mesh+Port. Effectively an Envoy Cluster name
	DestinationName string
	SNI             string
}

func BuildMeshDestinations(
	availableServices []*mesh_proto.ZoneIngress_AvailableService, // available services for a single mesh
	res xds_context.Resources,
	meshServices []*meshservice_api.MeshServiceResource,
	meshMzSvc []*v1alpha1.MeshMultiZoneServiceResource,
	mesServices []*meshexternalservice_api.MeshExternalServiceResource,
	systemNamespace string,
) MeshDestinations {
	return MeshDestinations{
		KumaIoServices: buildKumaIoServiceDestinations(availableServices, res),
		BackendRefs: append(
			append(buildMeshServiceDestinations(meshServices, systemNamespace), buildMeshMultiZoneServiceDestinations(meshMzSvc)...),
			buildMeshExternalServiceDestinations(mesServices)...,
		),
	}
}

func buildMeshServiceDestinations(
	meshServices []*meshservice_api.MeshServiceResource,
	systemNamespace string,
) []BackendRefDestination {
	var msDestinations []BackendRefDestination
	for _, ms := range meshServices {
		for _, port := range ms.Spec.Ports {
			sni := tls.SNIForResource(
				ms.SNIName(systemNamespace),
				ms.GetMeta().GetMesh(),
				meshservice_api.MeshServiceType,
				port.Port,
				nil,
			)
			msDestinations = append(msDestinations, BackendRefDestination{
				Mesh:            ms.GetMeta().GetMesh(),
				DestinationName: ms.DestinationName(port.Port),
				SNI:             sni,
			})
		}
	}
	return msDestinations
}

func buildMeshExternalServiceDestinations(
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
) []BackendRefDestination {
	var mesDestinations []BackendRefDestination
	for _, mes := range meshExternalServices {
		sni := tls.SNIForResource(
			core_model.GetDisplayName(mes.GetMeta()),
			mes.GetMeta().GetMesh(),
			meshexternalservice_api.MeshExternalServiceType,
			uint32(mes.Spec.Match.Port),
			nil,
		)
		mesDestinations = append(mesDestinations, BackendRefDestination{
			Mesh:            mes.GetMeta().GetMesh(),
			DestinationName: mes.DestinationName(uint32(mes.Spec.Match.Port)),
			SNI:             sni,
		})
	}
	return mesDestinations
}

func buildMeshMultiZoneServiceDestinations(
	meshMzSvc []*v1alpha1.MeshMultiZoneServiceResource,
) []BackendRefDestination {
	var msDestinations []BackendRefDestination
	for _, ms := range meshMzSvc {
		for _, port := range ms.Spec.Ports {
			msDestinations = append(msDestinations, BackendRefDestination{
				Mesh:            ms.GetMeta().GetMesh(),
				DestinationName: ms.DestinationName(port.Port),
				SNI: tls.SNIForResource(
					core_model.GetDisplayName(ms.GetMeta()),
					ms.GetMeta().GetMesh(),
					v1alpha1.MeshMultiZoneServiceType,
					port.Port,
					nil,
				),
			})
		}
	}
	return msDestinations
}

func buildKumaIoServiceDestinations(
	availableServices []*mesh_proto.ZoneIngress_AvailableService, // available services for a single mesh
	res xds_context.Resources,
) map[string][]envoy_tags.Tags {
	destForMesh := map[string][]envoy_tags.Tags{}
	trafficRoutes := res.TrafficRoutes().Items
	addTrafficRouteDestinations(trafficRoutes, destForMesh)
	meshHTTPRoutes := res.ListOrEmpty(meshhttproute_api.MeshHTTPRouteType).(*meshhttproute_api.MeshHTTPRouteResourceList).Items
	meshTCPRoutes := res.ListOrEmpty(meshtcproute_api.MeshTCPRouteType).(*meshtcproute_api.MeshTCPRouteResourceList).Items
	addMeshHTTPRouteDestinations(trafficRoutes, meshHTTPRoutes, destForMesh)
	addMeshTCPRouteDestinations(trafficRoutes, meshTCPRoutes, destForMesh)
	addGatewayRouteDestinations(res.GatewayRoutes().Items, destForMesh)
	addMeshGatewayDestinations(res.MeshGateways().Items, destForMesh)
	addVirtualOutboundDestinations(res.VirtualOutbounds().Items, availableServices, destForMesh)
	return destForMesh
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
		addDestination(backend.Destination, destinations)
	}
}

func addTrafficRouteDestinations(
	policies []*core_mesh.TrafficRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	for _, policy := range policies {
		for _, split := range policy.Spec.Conf.GetSplitWithDestination() {
			addDestination(split.Destination, destinations)
		}

		for _, http := range policy.Spec.Conf.Http {
			for _, split := range http.GetSplitWithDestination() {
				addDestination(split.Destination, destinations)
			}
		}
	}
}

func addMeshHTTPRouteDestinations(
	trafficRoutes []*core_mesh.TrafficRouteResource,
	policies []*meshhttproute_api.MeshHTTPRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	if len(trafficRoutes) == 0 {
		addTrafficFlowByDefaultDestination(destinations)
	}

	// Note that we're not merging these resources, but that's OK because the
	// set of destinations after merging is a subset of the set we get here by
	// iterating through them.
	for _, policy := range policies {
		for _, to := range policy.Spec.To {
			if toTags, ok := tags.FromTargetRef(to.TargetRef); ok {
				addMeshHTTPRouteToDestinations(to.Rules, toTags, destinations)
			}
		}
	}
}

func addMeshTCPRouteDestinations(
	trafficRoutes []*core_mesh.TrafficRouteResource,
	policies []*meshtcproute_api.MeshTCPRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	if len(trafficRoutes) == 0 {
		addTrafficFlowByDefaultDestination(destinations)
	}

	// Note that we're not merging these resources, but that's OK because the
	// set of destinations after merging is a subset of the set we get here by
	// iterating through them.
	for _, policy := range policies {
		for _, to := range policy.Spec.To {
			if toTags, ok := tags.FromTargetRef(to.TargetRef); ok {
				addMeshTCPRouteToDestinations(to.Rules, toTags, destinations)
			}
		}
	}
}

func addMeshHTTPRouteToDestinations(
	rules []meshhttproute_api.Rule,
	toTags envoy_tags.Tags,
	destinations map[string][]envoy_tags.Tags,
) {
	for _, rule := range rules {
		if rule.Default.BackendRefs == nil {
			addDestination(toTags, destinations)
			continue
		}

		for _, backendRef := range pointer.Deref(rule.Default.BackendRefs) {
			if tags, ok := tags.FromTargetRef(backendRef.TargetRef); ok {
				addDestination(tags, destinations)
			}
		}
	}
}

func addMeshTCPRouteToDestinations(
	rules []meshtcproute_api.Rule,
	toTags envoy_tags.Tags,
	destinations map[string][]envoy_tags.Tags,
) {
	for _, rule := range rules {
		if rule.Default.BackendRefs == nil {
			addDestination(toTags, destinations)
			continue
		}

		for _, backendRef := range rule.Default.BackendRefs {
			if tags, ok := tags.FromTargetRef(backendRef.TargetRef); ok {
				addDestination(tags, destinations)
			}
		}
	}
}

func addDestination(tags map[string]string, destinations map[string][]envoy_tags.Tags) {
	service := tags[mesh_proto.ServiceTag]
	destinations[service] = append(destinations[service], tags)
}

// addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist makes sure that when
// at least one MeshHTTPRoute policy exists there will be a "match all"
// destination pointing to all services (kuma.io/service:* -> kuma.io/service:*)
// This logic is necessary because of conflicting behaviors of TrafficRoute and
// MeshHTTPRoute policies. TrafficRoute expects that by default traffic doesn't
// flow, and there is necessary TrafficRoute with appropriate configuration
// to make communication between services possible. MeshHTTPRoute on the other
// hand expects the traffic to flow by default. As a result, when there is
// at least one MeshHTTPRoute policy present, traffic between services will flow
// by default, when there is none, it will flow, when appropriate TrafficRoute
// policy will exist.
func addTrafficFlowByDefaultDestination(
	destinations map[string][]envoy_tags.Tags,
) {
	// We need to add a destination to route any service to any instance of
	// that service
	matchAllTags := envoy_tags.Tags{mesh_proto.ServiceTag: mesh_proto.MatchAllTag}
	matchAllDestinations := destinations[mesh_proto.MatchAllTag]
	foundAllServicesDestination := slices.ContainsFunc(
		matchAllDestinations,
		func(tagsElem envoy_tags.Tags) bool {
			return reflect.DeepEqual(tagsElem, matchAllTags)
		},
	)

	if !foundAllServicesDestination {
		matchAllDestinations = append(matchAllDestinations, matchAllTags)
	}

	destinations[mesh_proto.MatchAllTag] = matchAllDestinations
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
