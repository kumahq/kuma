package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var RetryWsDefinition = ResourceWsDefinition{
	Name: "Retry",
	Path: "retries",
	ResourceFactory: func() model.Resource {
		return mesh.NewRetryResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.RetryResourceList{}
	},
}
