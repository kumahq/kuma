package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var SecretWsDefinition = ResourceWsDefinition{
	Name: "Secret",
	Path: "secrets",
	ResourceFactory: func() model.Resource {
		return system.NewSecretResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.SecretResourceList{}
	},
	Admin: true,
}
