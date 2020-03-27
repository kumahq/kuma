package mesh

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func (m *MeshResource) HasBuiltinCA() bool {
	switch m.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Builtin_:
		return true
	default:
		return false
	}
}

func (m *MeshResource) HasProvidedCA() bool {
	switch m.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Provided_:
		return true
	default:
		return false
	}
}

func (m *MeshResource) HasVaultCA() bool {
	switch m.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Vault_:
		return true
	default:
		return false
	}
}

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
