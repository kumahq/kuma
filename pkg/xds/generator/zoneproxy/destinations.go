package zoneproxy

import (
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_sni "github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

type MeshDestinations struct {
	BackendRefs []BackendRefDestination
}

type BackendRefDestination struct {
	resolve.ResolvedBackendRef

	SNI            string
	EndpointMapKey string
}

// BuildRealResourceDestinations produces one BackendRefDestination per
// (destination, port) pair.
func BuildRealResourceDestinations(destinations []core_resources.Destination) []BackendRefDestination {
	var result []BackendRefDestination
	for _, dest := range destinations {
		origin := kri.From(dest)

		for _, port := range dest.GetPorts() {
			id := kri.WithSectionName(origin, port.GetName())
			result = append(result, BackendRefDestination{
				SNI:            core_sni.FromKRI(id),
				EndpointMapKey: destinationname.MustResolve(dest, port),
				Ref: &resolve.RealResourceBackendRef{
					Resource: id,
					Origin:   origin,
					Weight:   1,
				},
			})
		}
	}
	return result
}

// IngressDestinations returns the destinations a zone ingress listener exposes:
// mesh services local to this zone, plus every MeshMultiZoneService.
func IngressDestinations(resources xds_context.Resources) []core_resources.Destination {
	var destinations []core_resources.Destination
	for _, ms := range resources.MeshServices().GetItems() {
		if ms.(*meshservice_api.MeshServiceResource).IsLocalMeshService() {
			destinations = append(destinations, ms.(core_resources.Destination))
		}
	}
	return append(destinations, resources.MeshMultiZoneServices().GetDestinations()...)
}

// EgressDestinations returns the destinations a zone egress listener proxies to.
func EgressDestinations(resources xds_context.Resources) []core_resources.Destination {
	return resources.MeshExternalServices().GetDestinations()
}
