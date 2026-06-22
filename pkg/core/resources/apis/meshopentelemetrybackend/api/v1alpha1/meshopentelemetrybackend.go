// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
)

// MeshOpenTelemetryBackend defines a shared OTel collector endpoint for observability policies.
// An empty spec is valid and represents the node-local default flow
// (kuma-dp resolves the address at runtime using HOST_IP or 127.0.0.1).
// +kuma:policy:is_policy=false
// +kuma:policy:allowed_on_system_namespace_only=true
// +kuma:policy:has_status=true
// +kuma:policy:short_name=motb
// +kuma:policy:kds_flags=model.GlobalToZonesFlag | model.ZoneToGlobalFlag
// +kuma:policy:singular_display_name=Mesh OpenTelemetry Backend
type MeshOpenTelemetryBackend struct {
	// Endpoint optionally defines the OTel collector address and port.
	// When omitted, the CP defaults port to 4317 and leaves address empty;
	// kuma-dp resolves the address at runtime using HOST_IP or 127.0.0.1.
	// +kubebuilder:validation:Optional
	Endpoint *Endpoint `json:"endpoint,omitempty"`
	// Protocol selects gRPC or HTTP transport for the collector connection.
	// Defaults to grpc when omitted.
	// +kubebuilder:validation:Optional
	Protocol *Protocol `json:"protocol,omitempty"`
	// Env controls whether standard OTEL exporter env vars participate in the
	// final exporter config for this backend.
	// Defaults to mode: Optional, precedence: EnvFirst, allowSignalOverrides: true
	// when omitted.
	// +kubebuilder:validation:Optional
	Env *EnvPolicy `json:"env,omitempty"`
}

// +kubebuilder:validation:Enum=grpc;http
type Protocol string

const (
	ProtocolGRPC Protocol = "grpc"
	ProtocolHTTP Protocol = "http"
)

// +kubebuilder:validation:Enum=Disabled;Optional;Required
type EnvMode string

const (
	EnvModeDisabled EnvMode = "Disabled"
	EnvModeOptional EnvMode = "Optional"
	EnvModeRequired EnvMode = "Required"
)

// +kubebuilder:validation:Enum=ExplicitFirst;EnvFirst
type EnvPrecedence string

const (
	EnvPrecedenceExplicitFirst EnvPrecedence = "ExplicitFirst"
	EnvPrecedenceEnvFirst      EnvPrecedence = "EnvFirst"
)

const (
	DefaultEnvMode              EnvMode       = EnvModeOptional
	DefaultEnvPrecedence        EnvPrecedence = EnvPrecedenceEnvFirst
	DefaultAllowSignalOverrides               = true
)

type EnvPolicy struct {
	// Mode controls whether OTEL env vars participate in the merge.
	// Disabled: env vars are skipped entirely; only explicit backend fields and
	// built-in defaults apply.
	// Optional (default): env vars are used when present; absence is fine.
	// Required: env vars must supply the missing fields - if any required field
	// is missing the signal is blocked (state: missing, RequiredEnvMissing in
	// blockedReasons) even when an explicit value or default could fill it.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=Optional
	Mode EnvMode `json:"mode"`
	// Precedence controls which source wins when both an explicit backend field
	// and an env var are present for the same field.
	// EnvFirst (default): env vars win; explicit backend fields fill gaps.
	// ExplicitFirst: explicit backend fields win; env vars fill gaps.
	// In either case, built-in defaults are the last fallback.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=EnvFirst
	Precedence EnvPrecedence `json:"precedence"`
	// AllowSignalOverrides controls whether per-signal OTEL env vars
	// (OTEL_EXPORTER_OTLP_TRACES_*, OTEL_EXPORTER_OTLP_METRICS_*,
	// OTEL_EXPORTER_OTLP_LOGS_*) may diverge from the shared
	// OTEL_EXPORTER_OTLP_* values.
	// true (default): per-signal vars override the shared values for that
	// signal.
	// false: per-signal vars are ignored; the shared values apply to all
	// signals. When per-signal overrides are dropped this way,
	// SignalOverridesDisallowed appears in blockedReasons (a soft block -
	// export still works via the shared config).
	AllowSignalOverrides *bool `json:"allowSignalOverrides,omitempty"`
}

type Endpoint struct {
	// Address of the OTel collector (hostname or IP).
	// When omitted, kuma-dp resolves it at runtime using HOST_IP or 127.0.0.1.
	// +kubebuilder:validation:Optional
	Address *string `json:"address,omitempty"`
	// Port of the OTel collector. Defaults to 4317 when omitted.
	// +kubebuilder:validation:Optional
	Port *int32 `json:"port,omitempty"`
	// Path is an optional base path prefix for HTTP endpoints.
	// The CP appends signal-specific suffixes (/v1/traces, /v1/metrics, /v1/logs).
	// Non-empty value is rejected by validation when protocol is grpc.
	Path *string `json:"path,omitempty"`
}

type MeshOpenTelemetryBackendStatus struct {
	Conditions []common_api.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (e *EnvPolicy) EffectiveMode() EnvMode {
	if e == nil || e.Mode == "" {
		return DefaultEnvMode
	}
	return e.Mode
}

func (e *EnvPolicy) EffectivePrecedence() EnvPrecedence {
	if e == nil || e.Precedence == "" {
		return DefaultEnvPrecedence
	}
	return e.Precedence
}

func (e *EnvPolicy) EffectiveAllowSignalOverrides() bool {
	if e == nil || e.AllowSignalOverrides == nil {
		return DefaultAllowSignalOverrides
	}
	return *e.AllowSignalOverrides
}

// Condition types
const (
	// ReferencedByPoliciesCondition indicates whether any observability policies reference this backend
	ReferencedByPoliciesCondition string = "ReferencedByPolicies"
)

// Condition reasons
const (
	// ReferencedReason indicates that one or more policies reference this backend
	ReferencedReason string = "Referenced"
	// NotReferencedReason indicates that no policies reference this backend
	NotReferencedReason string = "NotReferenced"
)
