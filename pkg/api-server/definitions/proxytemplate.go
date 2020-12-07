package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var ProxyTemplateWsDefinition = ResourceWsDefinition{
	Name: "ProxyTemplate",
	Path: "proxytemplates",
	ResourceFactory: func() model.Resource {
		return mesh.NewProxyTemplateResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.ProxyTemplateResourceList{}
	},
}
