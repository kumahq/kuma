package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var CircuitBreakerWsDefinition = ResourceWsDefinition{
	Name: "Circuit Breaker",
	Path: "circuit-breakers",
	ResourceFactory: func() model.Resource {
		return &mesh.CircuitBreakerResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.CircuitBreakerResourceList{}
	},
}
