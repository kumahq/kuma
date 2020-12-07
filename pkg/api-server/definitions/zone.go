package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ZoneWsDefinition = ResourceWsDefinition{
	Name: "Zone",
	Path: "zones",
	ResourceFactory: func() model.Resource {
		return system.NewZoneResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.ZoneResourceList{}
	},
}
