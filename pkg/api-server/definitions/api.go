package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
	ReadOnly            bool
	Admin               bool
}

var All = append(systemWsDefinitions, append(meshWsDefinitions, ResourceWsDefinition{
	Name: "GlobalSecret",
	Path: "global-secrets",
	ResourceFactory: func() model.Resource {
		return system.NewGlobalSecretResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &system.GlobalSecretResourceList{}
	},
	Admin:    true,
	ReadOnly: false,
})...)

func AllApis() core_rest.Api {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range All {
		resourceType := ws.ResourceFactory().GetType()
		mapping[resourceType] = core_rest.NewResourceApi(resourceType, ws.Path)
	}
	return &core_rest.ApiDescriptor{
		Resources: mapping,
	}
}
