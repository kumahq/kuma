package mesh

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
)

func (mesh *MeshResource) Default() {
	// default CA
	if mesh.Spec.Mtls == nil {
		mesh.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	if mesh.Spec.Mtls.Ca == nil {
		mesh.Spec.Mtls.Ca = &mesh_proto.CertificateAuthority{}
	}
	if mesh.Spec.Mtls.Ca.Type == nil {
		mesh.Spec.Mtls.Ca.Type = &mesh_proto.CertificateAuthority_Builtin_{
			Builtin: &mesh_proto.CertificateAuthority_Builtin{},
		}
	}
}
