package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var TimeoutWsDefinition = ResourceWsDefinition{
	Type: mesh.TimeoutType,
	Path: "timeouts",
}
