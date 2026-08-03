package zoneproxy

import (
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_resources "github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_sni "github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/tls"
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
func BuildRealResourceDestinations(destinations []core_resources.Destination, systemNS string, useNewSNIFormat bool) []BackendRefDestination {
	var result []BackendRefDestination
	for _, dest := range destinations {
		origin := kri.From(dest)
		mesh := dest.GetMeta().GetMesh()

		sniFor := newSNIBuilder(dest, origin, mesh, systemNS, useNewSNIFormat)

		for _, port := range dest.GetPorts() {
			id := kri.WithSectionName(origin, port.GetName())
			result = append(result, BackendRefDestination{
				SNI:            sniFor(id, port),
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

func newSNIBuilder(dest core_resources.Destination, origin kri.Identifier, mesh, systemNS string, useNewSNIFormat bool) func(kri.Identifier, core_resources.Port) string {
	if useNewSNIFormat {
		return func(id kri.Identifier, _ core_resources.Port) string {
			return core_sni.FromKRI(id)
		}
	}
	var rName string
	switch r := any(dest).(type) {
	case *meshservice_api.MeshServiceResource:
		rName = r.SNIName(systemNS)
	default:
		rName = core_model.GetDisplayName(dest.GetMeta())
	}
	return func(_ kri.Identifier, port core_resources.Port) string {
		return tls.SNIForResource(rName, mesh, origin.ResourceType, port.GetValue(), nil)
	}
}
