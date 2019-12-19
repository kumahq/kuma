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
	// default settings for Prometheus metrics
	if mesh.Spec.Metrics != nil {
		if mesh.Spec.Metrics.Prometheus != nil {
			if mesh.Spec.Metrics.Prometheus.Port == 0 {
				mesh.Spec.Metrics.Prometheus.Port = 5670
			}
			if mesh.Spec.Metrics.Prometheus.Path == "" {
				mesh.Spec.Metrics.Prometheus.Path = "/metrics"
			}
		}
	}
}
