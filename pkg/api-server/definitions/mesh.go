package definitions

import "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

var MeshWsDefinition = ResourceWsDefinition{
	Type: mesh.MeshType,
	Path: "meshes",
}
