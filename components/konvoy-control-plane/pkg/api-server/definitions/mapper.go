package definitions

import (
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_rest "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
)

func DefaultRESTMapper() core_rest.Mapper {
	return NewRESTMapper(MeshWsDefinition)
}

func NewRESTMapper(wss ...ResourceWsDefinition) core_rest.Mapper {
	mapping := make(map[core_model.ResourceType]core_rest.ResourceMapping)
	for _, ws := range wss {
		resourceType := ws.ResourceFactory().GetType()
		mapping[resourceType] = core_rest.ResourceMapping{CollectionPath: ws.Path}
	}
	return &core_rest.SimpleMapper{
		Resources: mapping,
	}
}
