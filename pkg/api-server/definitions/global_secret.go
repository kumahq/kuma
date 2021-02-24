package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var GlobalSecretWsDefinition = ResourceWsDefinition{
	Name: "GlobalSecret",
	Path: "global-secrets",
	ResourceFactory: func() model.Resource {
		return system.NewGlobalSecretResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.GlobalSecretResourceList{}
	},
	Admin: true,
}
