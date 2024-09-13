package delegated

import mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

type Config struct {
	Namespace                   string
	NamespaceOutsideMesh        string
	Mesh                        string
	KicIP                       string
	CpNamespace                 string
	ObservabilityDeploymentName string
	IPV6                        bool
	MeshServiceEnabled          mesh_proto.Mesh_MeshServices_Enabled
	UseEgress                   bool
}
