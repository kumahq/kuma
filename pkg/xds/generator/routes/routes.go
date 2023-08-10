package routes

import (
	"reflect"

	"golang.org/x/exp/slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func AddDestination(tags map[string]string, destinations map[string][]envoy_tags.Tags) {
	service := tags[mesh_proto.ServiceTag]
	destinations[service] = append(destinations[service], tags)
}

func AddMeshHTTPRouteDestinations(
	policies []*meshhttproute_api.MeshHTTPRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist(policies, destinations)

	// Note that we're not merging these resources, but that's OK because the
	// set of destinations after merging is a subset of the set we get here by
	// iterating through them.
	for _, policy := range policies {
		for _, to := range policy.Spec.To {
			if toTags, ok := to.TargetRef.EnvoyTags(); ok {
				addMeshHTTPRouteToDestinations(to.Rules, toTags, destinations)
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
			AddDestination(toTags, destinations)
			continue
		}

		for _, backendRef := range pointer.Deref(rule.Default.BackendRefs) {
			if tags, ok := backendRef.TargetRef.EnvoyTags(); ok {
				AddDestination(tags, destinations)
			}
		}
	}
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
func addTrafficFlowByDefaultDestinationIfMeshHTTPRoutesExist(
	policies []*meshhttproute_api.MeshHTTPRouteResource,
	destinations map[string][]envoy_tags.Tags,
) {
	// If there are no MeshHTTPRoutes, we are not modifying destinations
	if len(policies) == 0 {
		return
	}

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
