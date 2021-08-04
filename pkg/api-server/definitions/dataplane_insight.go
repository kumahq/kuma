package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var DataplaneInsightWsDefinition = ResourceWsDefinition{
	Type: mesh.DataplaneInsightType,
	Path: "dataplane-insights",
}
