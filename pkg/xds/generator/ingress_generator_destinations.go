package generator

import (
	"reflect"

	"golang.org/x/exp/slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func buildDestinations(
	ingressProxy *core_xds.ZoneIngressProxy,
) map[string][]envoy_tags.Tags {
	destinations := map[string][]envoy_tags.Tags{}
	policies := ingressProxy.PolicyResources

	for _, tr := range policies[core_mesh.TrafficRouteType].(*core_mesh.TrafficRouteResourceList).Items {
		for _, split := range tr.Spec.Conf.GetSplitWithDestination() {
			service := split.Destination[mesh_proto.ServiceTag]
			destinations[service] = append(destinations[service], split.Destination)
		}
		for _, http := range tr.Spec.Conf.Http {
			for _, split := range http.GetSplitWithDestination() {
				service := split.Destination[mesh_proto.ServiceTag]
				destinations[service] = append(destinations[service], split.Destination)
			}
		}
	}

	addMeshHTTPRoutesDestinations(policies[meshhttproute_api.MeshHTTPRouteType],
		destinations)

	var backends []*mesh_proto.MeshGatewayRoute_Backend

	for _, route := range ingressProxy.GatewayRoutes.Items {
		for _, rule := range route.Spec.GetConf().GetHttp().GetRules() {
			backends = append(backends, rule.Backends...)
		}
		for _, rule := range route.Spec.GetConf().GetTcp().GetRules() {
			backends = append(backends, rule.Backends...)
		}
	}

	for _, backend := range backends {
		service := backend.Destination[mesh_proto.ServiceTag]
		destinations[service] = append(destinations[service], backend.Destination)
	}

	for _, gateway := range ingressProxy.MeshGateways.Items {
		for _, selector := range gateway.Selectors() {
			service := selector.GetMatch()[mesh_proto.ServiceTag]
			for _, listener := range gateway.Spec.GetConf().GetListeners() {
				if !listener.CrossMesh {
					continue
				}
				destinations[service] = append(
					destinations[service],
					envoy_tags.Tags(mesh_proto.Merge(selector.GetMatch(), gateway.Spec.GetTags(), listener.GetTags())),
				)
			}
		}
	}

	return destinations
}

func addMeshHTTPRoutesDestinations(
	policyResources core_model.ResourceList,
	destinations map[string][]envoy_tags.Tags,
) {
	addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist(policyResources,
		destinations)

	policies := policyResources.(*meshhttproute_api.MeshHTTPRouteResourceList).
		Items

	// Note that we're not merging these resources, but that's OK because the
	// set of destinations after merging is a subset of the set we get here by
	// iterating through them.
	for _, policy := range policies {
		for _, to := range policy.Spec.To {
			toTags, ok := tagsFromTargetRef(to.TargetRef)
			if !ok {
				continue
			}

			for _, rule := range to.Rules {
				if rule.Default.BackendRefs == nil {
					service := toTags[mesh_proto.ServiceTag]
					destinations[service] = append(destinations[service],
						toTags)
				}

				backendRefs := pointer.Deref(rule.Default.BackendRefs)

				for _, backendRef := range backendRefs {
					backendTags, ok := tagsFromTargetRef(backendRef.TargetRef)
					if !ok {
						continue
					}

					service := backendTags[mesh_proto.ServiceTag]
					destinations[service] = append(destinations[service],
						backendTags)
				}
			}
		}
	}
}

// addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist makes sure that when
// at least one MeshHTTPRoute policy exists there will be a "match all"
// destination pointing to all services (kuma.io/service:* -> kuma.io/service:*)
// This logic is necessary because of conflicting behaviours of TrafficRoute and
// MeshHTTPRoute policies. TrafficRoute expects that by default traffic doesn't
// flow, and there is necessary TrafficRoute with appropriate configuration
// to make communication between services possible. MeshHTTPRoute on the other
// hand expects the traffic to flow by default. As a result, when there is
// at least one MeshHTTPRoute policy present, traffic between services will flow
// by default, when there is none, it will flow, when appropriate TrafficRoute
// policy will exist.
func addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist(
	policyResources core_model.ResourceList,
	destinations map[string][]envoy_tags.Tags,
) {
	if len(policyResources.GetItems()) > 0 {
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
}
