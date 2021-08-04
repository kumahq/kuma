package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var ExternalServiceWsDefinition = ResourceWsDefinition{
	Type: mesh.ExternalServiceType,
	Path: "external-services",
}
