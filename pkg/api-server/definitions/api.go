package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type ResourceWsDefinition struct {
	Type     model.ResourceType
	Path     string
	ReadOnly bool
	Admin    bool
}

func (ws *ResourceWsDefinition) ResourceFactory() model.Resource {
	m, err := registry.Global().NewObject(ws.Type)
	if err != nil {
		panic(err.Error())
	}

	return m
}

func (ws *ResourceWsDefinition) ResourceListFactory() model.ResourceList {
	l, err := registry.Global().NewList(ws.Type)
	if err != nil {
		panic(err.Error())
	}

	return l
}

var All []ResourceWsDefinition

func init() {
	All = append(All, systemWsDefinitions...)
	All = append(All, meshWsDefinitions...)
	All = append(All, ResourceWsDefinition{
		Type:     system.SecretType,
		Path:     "global-secrets",
		Admin:    true,
		ReadOnly: false,
	})
}

func AllApis() core_rest.Api {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range All {
		mapping[ws.Type] = core_rest.NewResourceApi(ws.Type, ws.Path)
	}
	return &core_rest.ApiDescriptor{
		Resources: mapping,
	}
}
