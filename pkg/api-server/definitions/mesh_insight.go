package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var MeshInsightWsDefinition = ResourceWsDefinition{
	Type: mesh.MeshType,
	Path: "mesh-insights",
}
