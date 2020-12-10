package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ZoneInsightWsDefinition = ResourceWsDefinition{
	Name: "Zone Insight",
	Path: "zone-insights",
	ResourceFactory: func() model.Resource {
		return system.NewZoneInsightResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.ZoneInsightResourceList{}
	},
}
