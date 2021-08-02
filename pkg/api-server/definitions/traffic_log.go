package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var TrafficLogWsDefinition = ResourceWsDefinition{
	Type: mesh.TrafficLogType,
	Path: "traffic-logs",
}
