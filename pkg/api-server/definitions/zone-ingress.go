package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ZoneIngressWsDefinition = ResourceWsDefinition{
	Name: "ZoneIngress",
	Path: "zone-ingresses",
	ResourceFactory: func() model.Resource {
		return mesh.NewZoneIngressResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.ZoneIngressResourceList{}
	},
}
