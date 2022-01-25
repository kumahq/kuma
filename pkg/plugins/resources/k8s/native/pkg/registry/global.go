package registry

import (
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

var global = NewTypeRegistry()

func Global() TypeRegistry {
	return global
}

func RegisterObjectType(typ ResourceType, obj model.KubernetesObject) {
	if err := global.RegisterObjectType(typ, obj); err != nil {
		panic(err)
	}
}

func RegisterObjectTypeIfAbsent(typ ResourceType, obj model.KubernetesObject) {
	global.RegisterObjectTypeIfAbsent(typ, obj)
}

func RegisterListType(typ ResourceType, obj model.KubernetesList) {
	if err := global.RegisterListType(typ, obj); err != nil {
		panic(err)
	}
}

func RegisterListTypeIfAbsent(typ ResourceType, obj model.KubernetesList) {
	global.RegisterListTypeIfAbsent(typ, obj)
}
