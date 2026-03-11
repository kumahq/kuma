package xds

import (
	"net"
	"path"
	"slices"
	"strconv"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
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
	Endpoint  *core_xds.Endpoint
	Protocol  motb_api.Protocol
	EnvPolicy *motb_api.EnvPolicy
	// UseHTTPS is true when OTLP/HTTP should use HTTPS transport.
	// Current heuristic: protocol=http with collector port 443.
	UseHTTPS bool
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
	backend, ambiguous := resolveBackendResourceByName(resources, name)
	if backend == nil {
		if ambiguous {
			otelLog.Info("MeshOpenTelemetryBackend reference is ambiguous, skipping backend", "name", name)
		} else {
			otelLog.Info("MeshOpenTelemetryBackend not found, skipping backend", "name", name)
		}
		return nil
	}

	spec := backend.Spec
	if spec.NodeEndpoint != nil {
		addr := nodeHostIP
		if addr == "" {
			addr = "127.0.0.1"
		}
		port := uint32(spec.NodeEndpoint.Port)
		return &ResolvedOtelBackend{
			Endpoint: &core_xds.Endpoint{
				Target: addr,
				Port:   port,
			},
			Protocol:  spec.Protocol,
			EnvPolicy: spec.Env,
			UseHTTPS:  shouldUseHTTPS(spec.Protocol, port),
			Path:      spec.NodeEndpoint.Path,
			Name:      name,
		}
	}
	port := uint32(spec.Endpoint.Port)
	return &ResolvedOtelBackend{
		Endpoint: &core_xds.Endpoint{
			Target: spec.Endpoint.Address,
			Port:   port,
		},
		Protocol:  spec.Protocol,
		EnvPolicy: spec.Env,
		UseHTTPS:  shouldUseHTTPS(spec.Protocol, port),
		Path:      spec.Endpoint.Path,
		Name:      name,
	}
}

func resolveBackendResourceByName(resources xds_context.Resources, name string) (*motb_api.MeshOpenTelemetryBackendResource, bool) {
	var direct *motb_api.MeshOpenTelemetryBackendResource
	var byDisplayName []*motb_api.MeshOpenTelemetryBackendResource

	for _, backend := range resources.MeshOpenTelemetryBackends().Items {
		if backend.GetMeta().GetName() == name {
			direct = backend
		}
		if backend.GetMeta().GetLabels()[mesh_proto.DisplayName] == name {
			byDisplayName = append(byDisplayName, backend)
		}
	}

	if direct != nil {
		return direct, false
	}

	switch len(byDisplayName) {
	case 0:
		return nil, false
	case 1:
		return byDisplayName[0], false
	default:
		return nil, true
	}
}

func shouldUseHTTPS(protocol motb_api.Protocol, port uint32) bool {
	return protocol == motb_api.ProtocolHTTP && port == 443
}

type AddResolvedBackendOptions struct {
	RefreshInterval string
}

// OtelBackendConfig is an alias for the unified OtelPipeBackend type.
// Kept for backward compatibility with per-signal dpapi packages.
type OtelBackendConfig = core_xds.OtelPipeBackend

func BuildResolvedPipeBackend(workDir string, resolved *ResolvedOtelBackend) core_xds.OtelPipeBackend {
	return core_xds.OtelPipeBackend{
		Name:       resolved.Name,
		SocketPath: core_xds.OpenTelemetrySocketName(workDir, resolved.Name),
		Endpoint:   CollectorEndpointString(resolved.Endpoint),
		UseHTTP:    resolved.Protocol == motb_api.ProtocolHTTP,
		UseHTTPS:   resolved.UseHTTPS,
		Path:       pointer.Deref(resolved.Path),
		EnvPolicy:  ResolveEnvPolicy(resolved.EnvPolicy),
	}
}

func ResolveEnvPolicy(policy *motb_api.EnvPolicy) core_xds.OtelResolvedEnvPolicy {
	return core_xds.OtelResolvedEnvPolicy{
		Mode:                 policy.EffectiveMode(),
		Precedence:           policy.EffectivePrecedence(),
		AllowSignalOverrides: policy.EffectiveAllowSignalOverrides(),
	}
}

func BuildSignalRuntimePlan(
	inventory *core_xds.OtelBootstrapInventory,
	envEnabled bool,
	envPolicy core_xds.OtelResolvedEnvPolicy,
	signal core_xds.OtelSignal,
	options AddResolvedBackendOptions,
) core_xds.OtelSignalRuntimePlan {
	plan := core_xds.OtelSignalRuntimePlan{
		Enabled:         true,
		RefreshInterval: options.RefreshInterval,
	}

	if inventory == nil {
		if !envEnabled && envPolicy.Mode != motb_api.EnvModeDisabled {
			plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonEnvDisabledByPlatform)
		}
		if envPolicy.Mode == motb_api.EnvModeRequired {
			plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonRequiredEnvMissing)
		}
		return plan
	}

	sharedInputPresent := inventory.Shared != nil && inventory.Shared.HasAnyInput()
	signalInventory := inventory.GetSignal(signal)
	signalInputPresent := signalInventory != nil && signalInventory.HasAnyInput()
	plan.EnvInputPresent = sharedInputPresent || signalInputPresent
	if signalInventory != nil {
		plan.OverrideKinds = slices.Clone(signalInventory.OverrideKinds)
	}

	if !envEnabled && envPolicy.Mode != motb_api.EnvModeDisabled {
		plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonEnvDisabledByPlatform)
	}
	if envPolicy.Mode == motb_api.EnvModeDisabled && plan.EnvInputPresent {
		plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonEnvDisabledByPolicy)
	}
	if envPolicy.Mode == motb_api.EnvModeRequired && !plan.EnvInputPresent {
		plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonRequiredEnvMissing)
	}
	if len(plan.OverrideKinds) > 0 && !envPolicy.AllowSignalOverrides {
		plan.BlockedReasons = append(plan.BlockedReasons, core_xds.OtelBlockedReasonSignalOverridesBlocked)
	}

	return plan
}

// AddResolvedToBackends adds a resolved OTel backend to the proxy accumulator.
// Shared by MeshTrace, MeshAccessLog, and MeshMetric plugins.
func AddResolvedToBackends(
	proxy *core_xds.Proxy,
	resolved *ResolvedOtelBackend,
	signal core_xds.OtelSignal,
	options AddResolvedBackendOptions,
) {
	base := BuildResolvedPipeBackend(proxy.Metadata.WorkDir, resolved)
	plan := BuildSignalRuntimePlan(
		proxy.Metadata.GetOtelEnvInventory(),
		proxy.Metadata.HasFeature(xds_types.FeatureOtelEnv),
		base.EnvPolicy,
		signal,
		options,
	)
	proxy.OtelPipeBackends.AddSignal(resolved.Name, base, signal, plan)
}

// EndpointForDirectOtelExport returns a copy of resolved endpoint adjusted for
// direct Envoy->collector transport.
func EndpointForDirectOtelExport(resolved *ResolvedOtelBackend) *core_xds.Endpoint {
	if resolved == nil || resolved.Endpoint == nil {
		return nil
	}

	ep := *resolved.Endpoint
	if resolved.UseHTTPS {
		if ep.ExternalService == nil {
			ep.ExternalService = &core_xds.ExternalService{}
		}
		ep.ExternalService.TLSEnabled = true
		ep.ExternalService.FallbackToSystemCa = true
	}

	return &ep
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
