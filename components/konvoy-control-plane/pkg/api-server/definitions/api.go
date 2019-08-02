package definitions

import (
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
)

func AllApis() core_rest.Api {
	return Apis(MeshWsDefinition, DataplaneWsDefinition, DataplaneStatusWsDefinition)
}

func Apis(wss ...ResourceWsDefinition) core_rest.Api {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceApi)
	for _, ws := range wss {
		resourceType := ws.ResourceFactory().GetType()
		mapping[resourceType] = core_rest.ResourceApi{CollectionPath: ws.Path}
	}
	return &core_rest.ApiDescriptor{
		Resources: mapping,
	}
}
