package zoneproxy

import (
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_sni "github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
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
				EndpointMapKey: destinationname.ResolveLegacyFromDestination(dest, port),
				ResolvedBackendRef: resolve.ResolvedBackendRef{
					Ref: &resolve.RealResourceBackendRef{
						Resource: id,
						Origin:   origin,
						Weight:   1,
					},
				},
			})
		}
	}
	return result
}
