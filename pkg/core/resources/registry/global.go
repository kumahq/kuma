package registry

import (
	"slices"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

var global = NewTypeRegistry()

func Global() TypeRegistry {
	return global
}

func RegisterType(res model.ResourceTypeDescriptor) {
	if err := global.RegisterType(res); err != nil {
		panic(err)
	}
}

func RegisterTypeIfAbsent(res model.ResourceTypeDescriptor) {
	if slices.Contains(global.ObjectTypes(), res.Name) {
		return
	}
	RegisterType(res)
}
