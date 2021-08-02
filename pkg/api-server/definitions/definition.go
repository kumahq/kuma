package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
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
