package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
)

var ZoneInsightWsDefinition = ResourceWsDefinition{
	Type: system.ZoneInsightType,
	Path: "zone-insights",
}
