package dpapi

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

const PATH = "/meshmetric"

type MeshMetricDpConfig struct {
	Observability Observability `json:"observability"`
}

type Observability struct {
	Metrics Metrics `json:"metrics"`
}

type Metrics struct {
	Applications []Application     `json:"applications"`
	Backends     []Backend         `json:"backends"`
	Sidecar      *v1alpha1.Sidecar `json:"sidecar,omitempty"`
	ExtraLabels  map[string]string `json:"extraLabels"`
}

type Application struct {
	Name    *string `json:"name,omitempty"`
	Path    string  `json:"path"`
	Port    uint32  `json:"port"`
	Address string  `json:"address"`
}

type Backend struct {
	Type          string                `json:"type"`
	Name          *string               `json:"name"`
	OpenTelemetry *OpenTelemetryBackend `json:"openTelemetry,omitempty"`
}

type OpenTelemetryBackend struct {
	Endpoint        string       `json:"endpoint"`
	RefreshInterval k8s.Duration `json:"refreshInterval"`
}
