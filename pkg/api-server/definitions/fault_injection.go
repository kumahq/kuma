package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var FaultInjectionWsDefinition = ResourceWsDefinition{
	Name: "Fault Injection",
	Path: "fault-injections",
	ResourceFactory: func() model.Resource {
		return &mesh.FaultInjectionResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.FaultInjectionResourceList{}
	},
}
