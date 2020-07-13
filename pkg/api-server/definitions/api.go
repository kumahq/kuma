package definitions

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rest "github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

func AllApis() core_rest.Api {
	return Apis(All...)
}

func Apis(wss ...ResourceWsDefinition) core_rest.Api {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range wss {
		resourceType := ws.ResourceFactory().GetType()
		mapping[resourceType] = core_rest.NewResourceApi(resourceType, ws.Path)
	}
	return &core_rest.ApiDescriptor{
		Resources: mapping,
	}
}
