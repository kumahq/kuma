package registry

import "github.com/kumahq/kuma/v2/pkg/core/resources/model"

var global = NewTypeRegistry()

func Global() TypeRegistry {
	return global
}

func RegisterType(res model.ResourceTypeDescriptor) {
	if err := global.RegisterType(res); err != nil {
		panic(err)
	}
}

func RegisterTypeValidator(res model.ResourceTypeDescriptor, validator AdditionalValidator) {
	global.RegisterValidator(res, validator)
}

func RegisterTypeIfAbsent(res model.ResourceTypeDescriptor) {
	for _, typ := range global.ObjectTypes() {
		if typ == res.Name {
			return
		}
	}
	RegisterType(res)
}
