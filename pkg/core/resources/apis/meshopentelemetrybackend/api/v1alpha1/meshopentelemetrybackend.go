// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
)

// MeshOpenTelemetryBackend defines a shared OTel collector endpoint for observability policies.
// +kuma:policy:is_policy=false
// +kuma:policy:allowed_on_system_namespace_only=true
// +kuma:policy:has_status=true
// +kuma:policy:short_name=motb
// +kuma:policy:kds_flags=model.GlobalToZonesFlag | model.ZoneToGlobalFlag
// +kuma:policy:singular_display_name=Mesh OpenTelemetry Backend
type MeshOpenTelemetryBackend struct {
	// Endpoint defines the OTel collector address and port.
	// Exactly one of endpoint or nodeEndpoint must be specified.
	// +kubebuilder:validation:Optional
	Endpoint *Endpoint `json:"endpoint,omitempty"`
	// NodeEndpoint connects to an OTel collector running as a DaemonSet
	// (hostPort mode) on the same node as the workload. The node's host IP
	// is injected by the injector and used as the target address.
	// Exactly one of endpoint or nodeEndpoint must be specified.
	// +kubebuilder:validation:Optional
	NodeEndpoint *NodeEndpoint `json:"nodeEndpoint,omitempty"`
	// Protocol selects gRPC or HTTP transport for the collector connection.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=grpc
	// +kubebuilder:validation:Enum=grpc;http
	Protocol Protocol `json:"protocol"`
	// Env controls whether standard OTEL exporter env vars participate in the
	// final exporter config for this backend.
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
	// Mode controls whether OTEL env vars are ignored, allowed, or required.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=Optional
	Mode EnvMode `json:"mode,omitempty"`
	// Precedence controls whether explicit backend fields or env vars win when
	// both are present for the same field.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=EnvFirst
	Precedence EnvPrecedence `json:"precedence,omitempty"`
	// AllowSignalOverrides controls whether signal-specific OTEL env vars such
	// as `OTEL_EXPORTER_OTLP_TRACES_*` may diverge from the shared config.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	AllowSignalOverrides *bool `json:"allowSignalOverrides,omitempty"`
}

// NodeEndpoint connects to an OTel collector running as a DaemonSet
// (hostPort mode). The node's host IP is used as the target address.
type NodeEndpoint struct {
	// Port of the OTel collector on the node.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
	// Path is an optional base path prefix for HTTP endpoints.
	// The CP appends signal-specific suffixes (/v1/traces, /v1/metrics, /v1/logs).
	// Non-empty value is rejected by validation when protocol is grpc.
	Path *string `json:"path,omitempty"`
}

type Endpoint struct {
	// Address of the OTel collector (hostname or IP).
	// +kubebuilder:validation:MinLength=1
	Address string `json:"address"`
	// Port of the OTel collector.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
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
