package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var ZoneWsDefinition = ResourceWsDefinition{
	Name: "Zone",
	Path: "zones",
	ResourceFactory: func() model.Resource {
		return &system.ZoneResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.ZoneResourceList{}
	},
}
