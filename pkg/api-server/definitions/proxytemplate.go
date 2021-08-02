package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

var ProxyTemplateWsDefinition = ResourceWsDefinition{
	Type: mesh.ProxyTemplateType,
	Path: "proxytemplates",
}
