package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// ResourceTypeForProxy returns the model resource type that implements
// the given mesh ProxyType.
func ResourceTypeForProxy(p mesh_proto.ProxyType) model.ResourceType {
	switch p {
	case mesh_proto.IngressProxyType:
		return ZoneIngressType
	case mesh_proto.GatewayProxyType:
		return DataplaneType
	default:
		return DataplaneType
	}
}
