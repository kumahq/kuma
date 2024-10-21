// +kubebuilder:object:generate=true
package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshAccessLog defines access log policies between different data plane
// proxies entities.
type MeshAccessLog struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
	// To list makes a match between the consumed services and corresponding configurations
	To []To `json:"to,omitempty"`
	// From list makes a match between clients and corresponding configurations
	From []From `json:"from,omitempty"`
}

type To struct {
	// TargetRef is a reference to the resource that represents a group of
	// destinations.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of destinations referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type From struct {
	// TargetRef is a reference to the resource that represents a group of
	// clients.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// Default is a configuration specific to the group of clients referenced in
	// 'targetRef'
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	Backends *[]Backend `json:"backends,omitempty"`
}

// +kubebuilder:validation:Enum=Tcp;File;OpenTelemetry
type BackendType string

const (
	TCPBackendType           BackendType = "Tcp"
	FileBackendType          BackendType = "File"
	OtelTelemetryBackendType BackendType = "OpenTelemetry"
)

type Backend struct {
	Type          BackendType  `json:"type"`
	Tcp           *TCPBackend  `json:"tcp,omitempty"`
	File          *FileBackend `json:"file,omitempty"`
	OpenTelemetry *OtelBackend `json:"openTelemetry,omitempty"`
}

// TCPBackend defines a TCP logging backend.
type TCPBackend struct {
	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators
	Format *Format `json:"format,omitempty"`
	// Address of the TCP logging backend
	// +kubebuilder:example="127.0.0.1:5000"
	// +kubebuilder:validation:MinLength=1
	Address string `json:"address"`
}

// Defines an OpenTelemetry logging backend.
type OtelBackend struct {
	// Attributes can contain placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators
	// +kubebuilder:example={{key: "mesh", value: "%KUMA_MESH%"}}
	Attributes []JsonValue `json:"attributes,omitempty"`
	// Body is a raw string or an OTLP any value as described at
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#field-body
	// It can contain placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators
	// +kubebuilder:example={kvlistValue: {values: {{key: "mesh", value: {stringValue: "%KUMA_MESH%"}}}}}
	Body *apiextensionsv1.JSON `json:"body,omitempty"`
	// Endpoint of OpenTelemetry collector. An empty port defaults to 4317.
	// +kubebuilder:example="otel-collector:4317"
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`
}

// FileBackend defines configuration for file based access logs
type FileBackend struct {
	// Format of access logs. Placeholders available on
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators
	Format *Format `json:"format,omitempty"`
	// Path to a file that logs will be written to
	// +kubebuilder:example="/tmp/access.log"
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`
}

// +kubebuilder:validation:Enum=Plain;Json
type FormatType string

const (
	PlainFormatType FormatType = "Plain"
	JsonFormatType  FormatType = "Json"
)

type Format struct {
	Type FormatType `json:"type"`
	// +kubebuilder:example="[%START_TIME%] %KUMA_MESH% %UPSTREAM_HOST%"
	Plain *string `json:"plain,omitempty"`
	// +kubebuilder:example={{key: "start_time", value: "%START_TIME%"},{key: "bytes_received", value: "%BYTES_RECEIVED%"}}
	Json *[]JsonValue `json:"json,omitempty"`
	// +kubebuilder:default=false
	OmitEmptyValues *bool `json:"omitEmptyValues,omitempty"`
}

type JsonValue struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
