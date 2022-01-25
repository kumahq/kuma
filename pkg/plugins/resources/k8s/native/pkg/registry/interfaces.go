package registry

import (
	"github.com/golang/protobuf/proto"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

type ResourceType = proto.Message

type TypeRegistry interface {
	RegisterObjectType(ResourceType, model.KubernetesObject) error
	RegisterObjectTypeIfAbsent(ResourceType, model.KubernetesObject)
	RegisterListType(ResourceType, model.KubernetesList) error
	RegisterListTypeIfAbsent(ResourceType, model.KubernetesList)

	NewObject(ResourceType) (model.KubernetesObject, error)
	NewList(ResourceType) (model.KubernetesList, error)
}
