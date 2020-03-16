package registry

import (
	"github.com/golang/protobuf/proto"

	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

type ResourceType = proto.Message

type TypeRegistry interface {
	RegisterObjectType(ResourceType, model.KubernetesObject) error
	RegisterListType(ResourceType, model.KubernetesList) error

	NewObject(ResourceType) (model.KubernetesObject, error)
	NewList(ResourceType) (model.KubernetesList, error)
}
