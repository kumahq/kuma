package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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
