package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var RateLimitWsDefinition = ResourceWsDefinition{
	Type: mesh.RateLimitType,
	Path: "rate-limits",
}
