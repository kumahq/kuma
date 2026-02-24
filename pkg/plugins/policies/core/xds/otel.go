package xds

import (

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

var otelLog = core.Log.WithName("otel-backend-resolution")

// ResolvedOtelBackend holds the resolved endpoint info from either a
// MeshOpenTelemetryBackend resource (via backendRef) or an inline endpoint.
type ResolvedOtelBackend struct {
	Endpoint *core_xds.Endpoint
	Protocol motb_api.Protocol
	// Path is the base path from MeshOpenTelemetryBackend (nil for inline endpoints and gRPC).
	Path *string
	// Name is used for naming clusters/listeners. For backendRef it's the resource name,
	// for inline endpoint it's derived from the endpoint string.
	Name string
}

// ResolveOtelBackend resolves a backendRef to a MeshOpenTelemetryBackend resource,
// falling back to the inline endpoint if backendRef is nil.
// Returns nil when the backendRef is dangling (resource not found).
func ResolveOtelBackend(
	backendRef *common_api.TargetRef,
	inlineEndpoint string,
	inlineEndpointParser func(string) *core_xds.Endpoint,
	inlineNameDeriver func(string) string,
	resources xds_context.Resources,
) *ResolvedOtelBackend {
	if backendRef != nil {
		return resolveFromBackendRef(backendRef, resources)
	}
	if inlineEndpoint != "" {
		return &ResolvedOtelBackend{
			Endpoint: inlineEndpointParser(inlineEndpoint),
			Protocol: motb_api.ProtocolGRPC,
			Name:     inlineNameDeriver(inlineEndpoint),
		}
	}
	return nil
}

func resolveFromBackendRef(ref *common_api.TargetRef, resources xds_context.Resources) *ResolvedOtelBackend {
	name := pointer.Deref(ref.Name)
	for _, backend := range resources.MeshOpenTelemetryBackends().Items {
		if backend.GetMeta().GetName() == name {
			spec := backend.Spec
			return &ResolvedOtelBackend{
				Endpoint: &core_xds.Endpoint{
					Target: spec.Endpoint.Address,
					Port:   uint32(spec.Endpoint.Port),
				},
				Protocol: spec.Protocol,
				Path:     spec.Endpoint.Path,
				Name:     name,
			}
		}
	}
	otelLog.Info("MeshOpenTelemetryBackend not found, skipping backend", "name", name)
	return nil
}
