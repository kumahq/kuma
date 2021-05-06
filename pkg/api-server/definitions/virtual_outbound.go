package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var VirtualOutboundWsDefinition = ResourceWsDefinition{
	Name: "Virtual Outbound",
	Path: "virtual-outbounds",
	ResourceFactory: func() model.Resource {
		return mesh.NewVirtualOutboundResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.VirtualOutboundResourceList{}
	},
}
