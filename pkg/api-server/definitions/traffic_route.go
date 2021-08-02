package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var TrafficRouteWsDefinition = ResourceWsDefinition{
	Type: mesh.TrafficRouteType,
	Path: "traffic-routes",
}
