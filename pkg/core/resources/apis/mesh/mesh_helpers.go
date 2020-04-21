package mesh

import (
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func (m *MeshResource) HasPrometheusMetricsEnabled() bool {
	return m != nil && m.Spec.GetMetrics().GetPrometheus() != nil
}

func (m *MeshResource) GetTracingBackend(name string) *mesh_proto.TracingBackend {
	backends := map[string]*mesh_proto.TracingBackend{}
	for _, backend := range m.Spec.GetTracing().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[m.Spec.GetTracing().GetDefaultBackend()]
	}
	return backends[name]
}

// GetLoggingBackends will return logging backends as comma separated strings
// if empty return empty string
func (m *MeshResource) GetLoggingBackends() string {
	var backends []string
	for _, backend := range m.Spec.GetLogging().GetBackends() {
		backends = append(backends, backend.GetName())
	}
	return strings.Join(backends, ", ")
}

// GetTracingBackends will return tracing backends as comma separated strings
// if empty return empty string
func (m *MeshResource) GetTracingBackends() string {
	var backends []string
	for _, backend := range m.Spec.GetTracing().GetBackends() {
		backends = append(backends, backend.GetName())
	}
	return strings.Join(backends, ", ")
}

func (m *MeshResource) GetDefaultCertificateAuthorityBackend() *mesh_proto.CertificateAuthorityBackend {
	return m.GetCertificateAuthorityBackend("")
}

func (m *MeshResource) GetCertificateAuthorityBackend(name string) *mesh_proto.CertificateAuthorityBackend {
	backends := map[string]*mesh_proto.CertificateAuthorityBackend{}
	for _, backend := range m.Spec.GetMtls().GetBackends() {
		backends[backend.Name] = backend
	}
	if name == "" {
		return backends[m.Spec.GetMtls().GetDefaultBackend()]
	}
	return backends[name]
}
