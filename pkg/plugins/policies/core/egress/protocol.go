package util

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// this method should be only used for zone egress because context is empty. In other
// cases informations about protocol are available in MeshContext
func GetExternalServiceProtocol(externalService *core_mesh.ExternalServiceResource) core_meta.Protocol {
	if externalService == nil {
		return core_meta.ProtocolUnknown
	}
	return core_meta.ParseProtocol(externalService.Spec.GetTags()[mesh_proto.ProtocolTag])
}
