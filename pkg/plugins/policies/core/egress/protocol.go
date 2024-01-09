package util

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// this method should be only used for zone egress because context is empty. In other
// cases informations about protocol are available in MeshContext
func GetExternalServiceProtocol(externalService *core_mesh.ExternalServiceResource) core_mesh.Protocol {
	if externalService == nil {
		return core_mesh.ProtocolUnknown
	}
	return core_mesh.ParseProtocol(externalService.Spec.GetTags()[mesh_proto.ProtocolTag])
}
