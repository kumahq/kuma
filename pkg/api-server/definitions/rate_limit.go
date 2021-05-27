package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var RateLimitWsDefinition = ResourceWsDefinition{
	Name: "Rate Limit",
	Path: "rate-limit",
	ResourceFactory: func() model.Resource {
		return mesh.NewRateLimitResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.RateLimitResourceList{}
	},
}
