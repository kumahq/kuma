package definitions

import "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

var DataplaneWsDefinition = ResourceWsDefinition{
	Type: mesh.DataplaneType,
	Path: "dataplanes",
}
