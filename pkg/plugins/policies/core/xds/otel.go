package xds

import (
	"net"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

// OTLP/HTTP signal path suffixes per the OpenTelemetry Protocol specification.
const (
	OtelTracesPathSuffix  = "v1/traces"
	OtelMetricsPathSuffix = "v1/metrics"
	OtelLogsPathSuffix    = "v1/logs"

	defaultOtelGrpcPort = 4317
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
func ResolveOtelBackend(
	backendRef *common_api.BackendResourceRef,
	inlineEndpoint string,
	inlineEndpointParser func(string) *core_xds.Endpoint,
	inlineNameDeriver func(string) string,
	resources xds_context.Resources,
) *ResolvedOtelBackend {
	switch {
	case backendRef != nil:
		return resolveFromBackendRef(backendRef, resources)
	case inlineEndpoint != "":
		return &ResolvedOtelBackend{
			Endpoint: inlineEndpointParser(inlineEndpoint),
			Protocol: motb_api.ProtocolGRPC,
			Name:     inlineNameDeriver(inlineEndpoint),
		}
	default:
		return nil
	}
}

func resolveFromBackendRef(ref *common_api.BackendResourceRef, resources xds_context.Resources) *ResolvedOtelBackend {
	var backend *motb_api.MeshOpenTelemetryBackendResource
	switch {
	case ref.Name != "":
		backend = resolveBackendResourceByName(resources, ref.Name)
	case len(ref.Labels) > 0:
		backend = resolveBackendResourceByLabels(resources, ref.Labels)
	}

	if backend == nil {
		otelLog.Info(
			"MeshOpenTelemetryBackend not found, skipping backend",
			"name", ref.Name,
			"labels", ref.Labels,
		)
		return nil
	}

	return resolvedFromSpec(backend)
}

func resolvedFromSpec(backend *motb_api.MeshOpenTelemetryBackendResource) *ResolvedOtelBackend {
	spec := backend.Spec
	protocol := pointer.DerefOr(spec.Protocol, motb_api.ProtocolGRPC)

	var addr string
	port := uint32(defaultOtelGrpcPort)
	var basePath *string
	if ep := spec.Endpoint; ep != nil {
		addr = pointer.Deref(ep.Address)
		if ep.Port != nil {
			port = uint32(*ep.Port)
		}
		basePath = ep.Path
	}

	return &ResolvedOtelBackend{
		Endpoint:  &core_xds.Endpoint{Target: addr, Port: port},
		Protocol:  protocol,
		EnvPolicy: spec.Env,
		UseHTTPS:  protocol == motb_api.ProtocolHTTP && port == 443,
		Path:      basePath,
		Name:      core_model.GetDisplayName(backend.GetMeta()),
	}
}

// resolveBackendResourceByName looks up by display name (the user-facing name
// without the namespace suffix that Kubernetes adds internally).
func resolveBackendResourceByName(
	resources xds_context.Resources,
	name string,
) *motb_api.MeshOpenTelemetryBackendResource {
	for _, backend := range resources.MeshOpenTelemetryBackends().Items {
		if core_model.GetDisplayName(backend.GetMeta()) == name {
			return backend
		}
	}
	return nil
}

// resolveBackendResourceByLabels matches all labels and picks the oldest on collision.
// Same strategy as DestinationIndex.resolveResourceIdentifier.
func resolveBackendResourceByLabels(
	resources xds_context.Resources,
	labels map[string]string,
) *motb_api.MeshOpenTelemetryBackendResource {
	selector := common_api.LabelSelector{MatchLabels: &labels}
	var matches []*motb_api.MeshOpenTelemetryBackendResource
	for _, backend := range resources.MeshOpenTelemetryBackends().Items {
		if selector.Matches(backend.GetMeta().GetLabels()) {
			matches = append(matches, backend)
		}
	}

	if len(matches) == 0 {
		return nil
	}

	oldest := matches[0]
	for _, b := range matches[1:] {
		if b.GetMeta().GetCreationTime().Before(oldest.GetMeta().GetCreationTime()) {
			oldest = b
		}
	}

	return oldest
}

type AddResolvedBackendOptions struct {
	RefreshInterval string
}

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

func ResolveEnvPolicy(policy *motb_api.EnvPolicy) *core_xds.OtelResolvedEnvPolicy {
	return &core_xds.OtelResolvedEnvPolicy{
		Mode:                 policy.EffectiveMode(),
		Precedence:           policy.EffectivePrecedence(),
		AllowSignalOverrides: policy.EffectiveAllowSignalOverrides(),
	}
}

func BuildSignalRuntimePlan(
	inventory *core_xds.OtelBootstrapInventory,
	envPolicy *core_xds.OtelResolvedEnvPolicy,
	signal core_xds.OtelSignal,
	options AddResolvedBackendOptions,
) core_xds.OtelSignalRuntimePlan {
	mode := motb_api.EnvModeOptional
	allowOverrides := false
	if envPolicy != nil {
		mode = envPolicy.Mode
		allowOverrides = envPolicy.AllowSignalOverrides
	}

	plan := core_xds.OtelSignalRuntimePlan{
		Enabled:         true,
		RefreshInterval: options.RefreshInterval,
	}

	// Collect env input state from the bootstrap inventory.
	if inventory != nil {
		signalInv := inventory.GetSignal(signal)

		sharedHasInput := inventory.Shared != nil && inventory.Shared.HasAnyInput()
		signalHasInput := signalInv != nil && signalInv.HasAnyInput()
		plan.EnvInputPresent = sharedHasInput || signalHasInput

		if signalInv != nil {
			plan.OverrideKinds = slices.Clone(signalInv.OverrideKinds)
		}

		if mode == motb_api.EnvModeRequired {
			plan.MissingFields = validationMissingFields(inventory, signal)
		}
	}

	// Determine blocked reasons based on mode vs actual env state.
	plan.BlockedReasons = blockedReasons(mode, allowOverrides, plan)

	return plan
}

func blockedReasons(
	mode motb_api.EnvMode,
	allowOverrides bool,
	plan core_xds.OtelSignalRuntimePlan,
) []string {
	var reasons []string

	switch mode {
	case motb_api.EnvModeDisabled:
		if plan.EnvInputPresent {
			reasons = append(reasons, core_xds.OtelBlockedReasonEnvDisabledByPolicy)
		}
	case motb_api.EnvModeRequired:
		if !plan.EnvInputPresent || len(plan.MissingFields) > 0 {
			reasons = append(reasons, core_xds.OtelBlockedReasonRequiredEnvMissing)
		}
	}

	if len(plan.OverrideKinds) > 0 && !allowOverrides {
		reasons = append(reasons, core_xds.OtelBlockedReasonSignalOverridesBlocked)
	}

	return reasons
}

func validationMissingFields(
	inventory *core_xds.OtelBootstrapInventory,
	signal core_xds.OtelSignal,
) []string {
	if inventory == nil {
		return nil
	}

	var missing []string
	for _, validationError := range inventory.ValidationErrors {
		scope, field, ok := strings.Cut(validationError, ".")
		if !ok {
			continue
		}

		if scope != "shared" && scope != string(signal) {
			continue
		}

		switch field {
		case "mtls":
			missing = appendUnique(missing, "client_certificate")
			missing = appendUnique(missing, "client_key")
		default:
			normalized := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(field), "-", "_"), " ", "_")
			missing = appendUnique(missing, normalized)
		}
	}

	return missing
}

func appendUnique(values []string, value string) []string {
	if value == "" || slices.Contains(values, value) {
		return values
	}
	return append(values, value)
}

// ResolveAddressForDirectExport fills in the target address when the MOTB spec
// left it empty (node-local default). Uses nodeHostIP from the proxy metadata,
// falling back to 127.0.0.1 for Universal mode.
func ResolveAddressForDirectExport(target, nodeHostIP string) string {
	switch {
	case target != "":
		return target
	case nodeHostIP != "":
		return nodeHostIP
	default:
		return "127.0.0.1"
	}
}

// EndpointForDirectOtelExport returns a copy of resolved endpoint adjusted for
// direct Envoy->collector transport. nodeHostIP fills in the address when the
// MOTB spec left it empty.
func EndpointForDirectOtelExport(resolved *ResolvedOtelBackend, nodeHostIP string) *core_xds.Endpoint {
	if resolved == nil || resolved.Endpoint == nil {
		return nil
	}

	ep := *resolved.Endpoint
	ep.Target = ResolveAddressForDirectExport(ep.Target, nodeHostIP)
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

	return net.JoinHostPort(
		endpoint.Target,
		strconv.FormatUint(uint64(endpoint.Port), 10),
	)
}

// ParseOtelEndpoint parses a "host:port" endpoint string into an Endpoint.
// Prepends a synthetic scheme so url.Parse handles IPv6 and port parsing.
// Defaults to gRPC port 4317 when no port is specified.
func ParseOtelEndpoint(endpoint string) *core_xds.Endpoint {
	u, err := url.Parse("otel://" + endpoint)
	if err != nil || u.Hostname() == "" {
		return &core_xds.Endpoint{Target: endpoint, Port: defaultOtelGrpcPort}
	}

	port := uint32(defaultOtelGrpcPort)
	if p := u.Port(); p != "" {
		if val, err := strconv.ParseUint(p, 10, 16); err == nil && val > 0 {
			port = uint32(val)
		}
	}

	return &core_xds.Endpoint{
		Target: u.Hostname(),
		Port:   port,
	}
}
