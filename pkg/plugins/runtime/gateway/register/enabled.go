package register

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

func RegisterGatewayTypes() {
	// RegisterTypeIfAbsent is used because it's not deterministic in testing that RegisterGatewayTypes is called only once.
	registry.RegisterTypeIfAbsent(core_mesh.MeshGatewayResourceTypeDescriptor)
	registry.RegisterTypeIfAbsent(core_mesh.MeshGatewayRouteResourceTypeDescriptor)
}
