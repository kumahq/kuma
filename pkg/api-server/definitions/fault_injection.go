package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var FaultInjectionWsDefinition = ResourceWsDefinition{
	Type: mesh.FaultInjectionType,
	Path: "fault-injections",
}
