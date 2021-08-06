package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

type ResourceWsDefinition struct {
	Type      core_model.ResourceType
	Path      string
	ReadOnly  bool
	AdminOnly bool
}

func (ws *ResourceWsDefinition) ResourceFactory() core_model.Resource {
	m, err := registry.Global().NewObject(ws.Type)
	if err != nil {
		panic(err.Error())
	}

	return m
}

func (ws *ResourceWsDefinition) ResourceListFactory() core_model.ResourceList {
	l, err := registry.Global().NewList(ws.Type)
	if err != nil {
		panic(err.Error())
	}

	return l
}

var All = append(append([]ResourceWsDefinition{
	{
		Type:      system.GlobalSecretType,
		Path:      "global-secrets",
		AdminOnly: true,
	},
}, systemWsDefinitions...), meshWsDefinitions...)

func AllApis() core_rest.Api {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range All {
		mapping[ws.Type] = core_rest.NewResourceApi(ws.Type, ws.Path)
	}
	return &core_rest.ApiDescriptor{
		Resources: mapping,
	}
}
