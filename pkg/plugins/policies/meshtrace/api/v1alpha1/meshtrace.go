// +kubebuilder:object:generate=true
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshTrace allows users to enable request tracing between services in the mesh
// and sending these traces to a third party storage.
type MeshTrace struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined inplace.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// MeshTrace configuration.
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	// A one element array of backend definition.
	// Envoy allows configuring only 1 backend, so the natural way of
	// representing that would be just one object. Unfortunately due to the
	// reasons explained in MADR 009-tracing-policy this has to be a one element
	// array for now.
	Backends *[]Backend `json:"backends,omitempty"`
	// Sampling configuration.
	// Sampling is the process by which a decision is made on whether to
	// process/export a span or not.
	Sampling *Sampling `json:"sampling,omitempty"`
	// Custom tags configuration. You can add custom tags to traces based on
	// headers or literal values.
	Tags *[]Tag `json:"tags,omitempty"`
}

// +kubebuilder:validation:Enum=Zipkin;Datadog;OpenTelemetry
type BackendType string

const (
	ZipkinBackendType        BackendType = "Zipkin"
	DatadogBackendType       BackendType = "Datadog"
	OpenTelemetryBackendType BackendType = "OpenTelemetry"
)

// Only one of zipkin, datadog or openTelemetry can be used.
type Backend struct {
	Type BackendType `json:"type"`
	// Zipkin backend configuration.
	Zipkin *ZipkinBackend `json:"zipkin,omitempty"`
	// Datadog backend configuration.
	Datadog *DatadogBackend `json:"datadog,omitempty"`
	// OpenTelemetry backend configuration.
	OpenTelemetry *OpenTelemetryBackend `json:"openTelemetry,omitempty"`
}

// OpenTelemetry tracing backend configuration.
type OpenTelemetryBackend struct {
	// Address of OpenTelemetry collector.
	// +kubebuilder:example="otel-collector:4317"
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`
}

// Zipkin tracing backend configuration.
type ZipkinBackend struct {
	// Address of Zipkin collector.
	Url string `json:"url"`
	// Generate 128bit traces. Default: false
	TraceId128Bit *bool `json:"traceId128bit,omitempty"`
	// Version of the API. values: httpJson, httpProto. Default:
	// httpJson see
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/trace/v3/zipkin.proto#L66
	// +kubebuilder:default="httpJson"
	// +kubebuilder:validation:Enum=httpJson;httpProto
	ApiVersion *string `json:"apiVersion,omitempty"`
	// Determines whether client and server spans will share the same span
	// context. Default: true.
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/trace/v3/zipkin.proto#L63
	SharedSpanContext *bool `json:"sharedSpanContext,omitempty"`
}

// Datadog tracing backend configuration.
type DatadogBackend struct {
	// Address of Datadog collector, only host and port are allowed (no paths,
	// fragments etc.)
	Url string `json:"url"`
	// Determines if datadog service name should be split based on traffic
	// direction and destination. For example, with `splitService: true` and a
	// `backend` service that communicates with a couple of databases, you would
	// get service names like `backend_INBOUND`, `backend_OUTBOUND_db1`, and
	// `backend_OUTBOUND_db2` in Datadog. Default: false
	SplitService *bool `json:"splitService,omitempty"`
}

// Sampling configuration.
type Sampling struct {
	// Target percentage of requests will be traced
	// after all other sampling checks have been applied (client, force tracing,
	// random sampling). This field functions as an upper limit on the total
<<<<<<< HEAD
	// configured sampling rate. For instance, setting client_sampling to 100%
	// but overall_sampling to 1% will result in only 1% of client requests with
	// the appropriate headers to be force traced. Default: 100% Mirror of
	// overall_sampling in Envoy
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#L142-L150
	// Either int or decimal represented as string.
=======
	// configured sampling rate. For instance, setting client to 100
	// but overall to 1 will result in only 1% of client requests with
	// the appropriate headers to be force traced. Mirror of
	// overall_sampling in Envoy
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#L142-L150
	// Either int or decimal represented as string.
	// +kubebuilder:default=100
>>>>>>> 79a23897c (fix(MeshTrace): invalid sampling default values (#11548))
	Overall *intstr.IntOrString `json:"overall,omitempty"`
	// Target percentage of requests that will be force traced if the
	// 'x-client-trace-id' header is set. Default: 100% Mirror of
	// client_sampling in Envoy
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#L127-L133
	// Either int or decimal represented as string.
<<<<<<< HEAD
=======
	// +kubebuilder:default=100
>>>>>>> 79a23897c (fix(MeshTrace): invalid sampling default values (#11548))
	Client *intstr.IntOrString `json:"client,omitempty"`
	// Target percentage of requests that will be randomly selected for trace
	// generation, if not requested by the client or not forced. Default: 100%
	// Mirror of random_sampling in Envoy
	// https://github.com/envoyproxy/envoy/blob/v1.22.0/api/envoy/config/filter/network/http_connection_manager/v2/http_connection_manager.proto#L135-L140
	// Either int or decimal represented as string.
<<<<<<< HEAD
=======
	// +kubebuilder:default=100
>>>>>>> 79a23897c (fix(MeshTrace): invalid sampling default values (#11548))
	Random *intstr.IntOrString `json:"random,omitempty"`
}

// Custom tags configuration.
// Only one of literal or header can be used.
type Tag struct {
	// Name of the tag.
	Name string `json:"name"`
	// Tag taken from literal value.
	Literal *string `json:"literal,omitempty"`
	// Tag taken from a header.
	Header *HeaderTag `json:"header,omitempty"`
}

// Tag taken from a header configuration.
type HeaderTag struct {
	// Name of the header.
	Name string `json:"name"`
	// Default value to use if header is missing.
	// If the default is missing and there is no value the tag will not be
	// included.
	Default *string `json:"default,omitempty"`
}
