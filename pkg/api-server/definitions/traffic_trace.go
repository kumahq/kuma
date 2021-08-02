package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var TrafficTraceWsDefinition = ResourceWsDefinition{
	Type: mesh.TrafficTraceType,
	Path: "traffic-traces",
}
