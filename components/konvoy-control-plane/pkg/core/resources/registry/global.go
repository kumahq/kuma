package registry

import "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

var global = NewTypeRegistry()

func Global() TypeRegistry {
	return global
}

func RegisterType(resType model.ResourceType, res model.Resource) {
	if err := global.RegisterType(resType, res); err != nil {
		panic(err)
	}
}

func RegistryListType(resType model.ResourceType, resList model.ResourceList) {
	if err := global.RegisterListType(resType, resList); err != nil {
		panic(err)
	}
}
