package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var TrafficPermissionWsDefinition = ResourceWsDefinition{
	Type: mesh.TrafficPermissionType,
	Path: "traffic-permissions",
}
