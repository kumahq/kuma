package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var ServiceInsightWsDefinition = ResourceWsDefinition{
	Type: mesh.ServiceInsightType,
	Path: "service-insights",
}
