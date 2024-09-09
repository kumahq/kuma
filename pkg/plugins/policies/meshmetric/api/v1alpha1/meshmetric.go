// +kubebuilder:object:generate=true
package v1alpha1

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

// MeshMetric allows users to enable metrics collection in the Mesh, and define third party backend for storing them.
type MeshMetric struct {
	// TargetRef is a reference to the resource the policy takes an effect on.
	// The resource could be either a real store object or virtual resource
	// defined in-place.
	TargetRef *common_api.TargetRef `json:"targetRef,omitempty"`
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
	// Profiles allows to customize which metrics are published.
	Profiles *Profiles `json:"profiles,omitempty"`
	// IncludeUnused if false will scrape only metrics that has been by sidecar (counters incremented
	// at least once, gauges changed at least once, and histograms added to at
	// least once). If true will scrape all metrics (even the ones with zeros).
	// +kubebuilder:default=false
	IncludeUnused *bool `json:"includeUnused,omitempty"`
}

type Profiles struct {
	// AppendProfiles allows to combine the metrics from multiple predefined profiles.
	AppendProfiles *[]Profile `json:"appendProfiles,omitempty"`
	// Exclude makes it possible to exclude groups of metrics from a resulting profile.
	// Exclude is subordinate to Include.
	Exclude *[]Selector `json:"exclude,omitempty"`
	// Include makes it possible to include additional metrics in a selected profiles.
	// Include takes precedence over Exclude.
	Include *[]Selector `json:"include,omitempty"`
}

type Profile struct {
	// Name of the predefined profile, one of: all, basic, none
	Name ProfileName `json:"name"`
}

// +kubebuilder:validation:Enum=All;Basic;None
type ProfileName string

const (
	AllProfileName   ProfileName = "All"
	BasicProfileName ProfileName = "Basic"
	NoneProfileName  ProfileName = "None"
)

type Selector struct {
	// Type defined the type of selector, one of: prefix, regex, exact
	Type SelectorType `json:"type"`
	// Match is the value used to match using particular Type
	Match string `json:"match"`
}

// +kubebuilder:validation:Enum=Prefix;Regex;Exact;Contains
type SelectorType string

const (
	PrefixSelectorType   SelectorType = "Prefix"
	RegexSelectorType    SelectorType = "Regex"
	ExactSelectorType    SelectorType = "Exact"
	ContainsSelectorType SelectorType = "Contains"
)

type Application struct {
	// Name of the application to scrape
	Name *string `json:"name,omitempty"`
	// Path on which an application expose HTTP endpoint with metrics.
	// +kubebuilder:default="/metrics/prometheus"
	Path *string `json:"path,omitempty"`
	// Address on which an application listens.
	Address *string `json:"address,omitempty"`
	// Port on which an application expose HTTP endpoint with metrics.
	Port uint32 `json:"port"`
}

type Backend struct {
	// Type of the backend that will be used to collect metrics. At the moment only Prometheus backend is available.
	Type BackendType `json:"type"`
	// Prometheus backend configuration.
	Prometheus *PrometheusBackend `json:"prometheus,omitempty"`
	// OpenTelemetry backend configuration
	OpenTelemetry *OpenTelemetryBackend `json:"openTelemetry,omitempty"`
}

// +kubebuilder:validation:Enum=Prometheus;OpenTelemetry
type BackendType string

const (
	PrometheusBackendType    BackendType = "Prometheus"
	OpenTelemetryBackendType BackendType = "OpenTelemetry"
)

type PrometheusBackend struct {
	// ClientId of the Prometheus backend. Needed when using MADS for DP discovery.
	ClientId *string `json:"clientId,omitempty"`
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

type OpenTelemetryBackend struct {
	// Endpoint for OpenTelemetry collector
	Endpoint string `json:"endpoint"`
	// RefreshInterval defines how frequent metrics should be pushed to collector
	RefreshInterval *k8s.Duration `json:"refreshInterval,omitempty"`
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
