package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var ZoneIngressWsDefinition = ResourceWsDefinition{
	Type: mesh.ZoneIngressType,
	Path: "zone-ingresses",
}
