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
