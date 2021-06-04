package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var RateLimitWsDefinition = ResourceWsDefinition{
	Name: "Rate Limit",
	Path: "rate-limits",
	ResourceFactory: func() model.Resource {
		return mesh.NewRateLimitResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.RateLimitResourceList{}
	},
}
