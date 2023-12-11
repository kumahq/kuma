// +kubebuilder:object:generate=true
package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshMetric allows users to enable metrics collection in the Mesh, and define third party backend for storing them.
type MeshMetric struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef common_api.TargetRef `json:"targetRef"`
	// MeshMetric configuration.
	Default Conf `json:"default,omitempty"`
}

type Conf struct {
	// Sidecar metrics collection configuration
	Sidecar *Sidecar `json:"sidecar,omitempty"`
	// Applications is a list of application that Dataplane Proxy will scrape
	Applications *[]Application `json:"applications,omitempty"`
	// Backends list that will be used to collect metrics.
	Backends *[]Backend `json:"backends,omitempty"`
}

type Sidecar struct {
	// Regex that will be used to filter sidecar metrics. It uses Google RE2 engine https://github.com/google/re2
	Regex *string `json:"regex,omitempty"`
	// UsedOnly will scrape only metrics that has been by sidecar (counters incremented
	// at least once, gauges changed at least once, and histograms added to at
	// least once).
	// +kubebuilder:default=false
	UsedOnly *bool `json:"usedOnly,omitempty"`
}

type Application struct {
	// Path on which an application expose HTTP endpoint with metrics.
	// +kubebuilder:default="/metrics/prometheus"
	Path *string `json:"path,omitempty"`
	// Port on which an application expose HTTP endpoint with metrics.
	Port uint32 `json:"port"`
}

type Backend struct {
	// Type of the backend that will be used to collect metrics. At the moment only Prometheus backend is available.
	Type BackendType `json:"type"`
	// Name of the backend. Needed when using MADS for DP discovery.
	Name *string `json:"name,omitempty"`
	// Prometheus backend configuration.
	Prometheus *PrometheusBackend `json:"prometheus,omitempty"`
}

// +kubebuilder:validation:Enum=Prometheus
type BackendType string

const PrometheusBackendType BackendType = "Prometheus"

type PrometheusBackend struct {
	// Port on which a dataplane should expose HTTP endpoint with Prometheus metrics.
	// +kubebuilder:default=5670
	Port uint32 `json:"port"`
	// Path on which a dataplane should expose HTTP endpoint with Prometheus metrics.
	// +kubebuilder:default="/metrics"
	Path string `json:"path"`
	// Configuration of TLS for prometheus listener.
	Tls *PrometheusTls `json:"tls,omitempty"`
}

type PrometheusTls struct {
	// Configuration of TLS for Prometheus listener.
	// +kubebuilder:default="Disabled"
	Mode TlsMode `json:"mode"`
}

// +kubebuilder:validation:Enum=Disabled;ProvidedTLS;ActiveMTLSBackend
type TlsMode string

const (
	// Disabled Tls for Prometheus listener
	Disabled TlsMode = "Disabled"
	// ProvidedTLS means that user is responsible for providing certificates to dataplanes.
	// Path for the certificate and the key needs to be provided to the dataplane
	// by environments variables:
	// * KUMA_DATAPLANE_RUNTIME_METRICS_CERT_PATH
	// * KUMA_DATAPLANE_RUNTIME_METRICS_KEY_PATH
	ProvidedTLS TlsMode = "ProvidedTLS"
	// ActiveMTLSBackend means that control-plane delivers certificates to the prometheus listener.
	// This should be used when prometheus is running inside the Mesh.
	ActiveMTLSBackend TlsMode = "ActiveMTLSBackend"
)
