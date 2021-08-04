package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var CircuitBreakerWsDefinition = ResourceWsDefinition{
	Type: mesh.CircuitBreakerType,
	Path: "circuit-breakers",
}
