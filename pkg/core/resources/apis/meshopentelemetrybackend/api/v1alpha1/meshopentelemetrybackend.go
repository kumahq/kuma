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
	Endpoint Endpoint `json:"endpoint"`
	// Protocol selects gRPC or HTTP transport for the collector connection.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=grpc
	// +kubebuilder:validation:Enum=grpc;http
	Protocol Protocol `json:"protocol"`
}

// +kubebuilder:validation:Enum=grpc;http
type Protocol string

const (
	ProtocolGRPC Protocol = "grpc"
	ProtocolHTTP Protocol = "http"
)

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
	// Ignored for gRPC.
	Path *string `json:"path,omitempty"`
}

type MeshOpenTelemetryBackendStatus struct {
	Conditions []common_api.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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
