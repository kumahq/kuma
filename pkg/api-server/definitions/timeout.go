package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var TimeoutWsDefinition = ResourceWsDefinition{
	Name: "Timeout",
	Path: "timeouts",
	ResourceFactory: func() model.Resource {
		return mesh.NewTimeoutResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TimeoutResourceList{}
	},
}
