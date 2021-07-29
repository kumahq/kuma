package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ZoneIngressInsightWsDefinition = ResourceWsDefinition{
	Name: "Zone Ingress Insight",
	Path: "zone-ingress-insights",
	ResourceFactory: func() model.Resource {
		return mesh.NewZoneIngressInsightResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.ZoneIngressInsightResourceList{}
	},
}
