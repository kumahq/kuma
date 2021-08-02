package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var HealthCheckWsDefinition = ResourceWsDefinition{
	Type: mesh.HealthCheckType,
	Path: "health-checks",
}
