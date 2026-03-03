package xds

import (
	"net"
	"path"
	"strconv"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

// OTLP/HTTP signal path suffixes per the OpenTelemetry Protocol specification.
const (
	OtelTracesPathSuffix  = "v1/traces"
	OtelMetricsPathSuffix = "v1/metrics"
	OtelLogsPathSuffix    = "v1/logs"
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

// FullPath joins the base path from MeshOpenTelemetryBackend with the given
// OTLP signal suffix (e.g. OtelTracesPathSuffix). Returns "/" + suffix when
// no base path is set.
func (r *ResolvedOtelBackend) FullPath(signalSuffix string) string {
	base := "/"
	if r.Path != nil {
		base = *r.Path
	}
	return path.Join(base, signalSuffix)
}

// ResolveOtelBackend resolves a backendRef to a MeshOpenTelemetryBackend resource,
// falling back to the inline endpoint if backendRef is nil.
// Returns nil when the backendRef is dangling (resource not found) or no config exists.
// nodeHostIP is the host IP of the node running the workload, used when the backend
// specifies nodeEndpoint. Falls back to 127.0.0.1 when empty (Universal mode).
func ResolveOtelBackend(
	backendRef *common_api.BackendResourceRef,
	inlineEndpoint string,
	inlineEndpointParser func(string) *core_xds.Endpoint,
	inlineNameDeriver func(string) string,
	resources xds_context.Resources,
	nodeHostIP string,
) *ResolvedOtelBackend {
	if backendRef != nil {
		return resolveFromBackendRef(backendRef, resources, nodeHostIP)
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

func resolveFromBackendRef(ref *common_api.BackendResourceRef, resources xds_context.Resources, nodeHostIP string) *ResolvedOtelBackend {
	name := ref.Name
	for _, backend := range resources.MeshOpenTelemetryBackends().Items {
		displayName := backend.GetMeta().GetLabels()[mesh_proto.DisplayName]
		if displayName == name || backend.GetMeta().GetName() == name {
			spec := backend.Spec
			if spec.NodeEndpoint != nil {
				addr := nodeHostIP
				if addr == "" {
					addr = "127.0.0.1"
				}
				return &ResolvedOtelBackend{
					Endpoint: &core_xds.Endpoint{
						Target: addr,
						Port:   uint32(spec.NodeEndpoint.Port),
					},
					Protocol: spec.Protocol,
					Path:     spec.NodeEndpoint.Path,
					Name:     name,
				}
			}
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

// OtelBackendConfig is an alias for the unified OtelPipeBackend type.
// Kept for backward compatibility with per-signal dpapi packages.
type OtelBackendConfig = core_xds.OtelPipeBackend

// AddResolvedToBackends adds a resolved OTel backend to the proxy accumulator.
// Shared by MeshTrace, MeshAccessLog, and MeshMetric plugins.
func AddResolvedToBackends(proxy *core_xds.Proxy, resolved *ResolvedOtelBackend) {
	proxy.OtelPipeBackends.Add(resolved.Name, core_xds.OtelPipeBackend{
		SocketPath: core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir, resolved.Name),
		Endpoint:   CollectorEndpointString(resolved.Endpoint),
		UseHTTP:    resolved.Protocol == motb_api.ProtocolHTTP,
		Path:       pointer.Deref(resolved.Path),
	})
}

// CollectorEndpointString formats an Endpoint as a "host:port" string suitable
// for dialing. IPv6 addresses are bracketed by net.JoinHostPort.
func CollectorEndpointString(endpoint *core_xds.Endpoint) string {
	if endpoint.Port == 0 {
		return endpoint.Target
	}
	return net.JoinHostPort(endpoint.Target, strconv.Itoa(int(endpoint.Port)))
}

// ParseOtelEndpoint parses a "host:port" endpoint string into an Endpoint.
// Handles IPv6 addresses (bracketed and bare) via net.SplitHostPort.
// Defaults to gRPC port 4317 when no port is present or parsing fails.
func ParseOtelEndpoint(endpoint string) *core_xds.Endpoint {
	host, portStr, err := net.SplitHostPort(endpoint)
	port := uint32(4317)
	if err == nil {
		if val, err := strconv.ParseInt(portStr, 10, 32); err == nil && val > 0 && val <= 65535 {
			port = uint32(val)
		}
	} else {
		host = endpoint
		if l := len(host); l > 1 && host[0] == '[' && host[l-1] == ']' {
			host = host[1 : l-1]
		}
	}
	return &core_xds.Endpoint{
		Target: host,
		Port:   port,
	}
}
