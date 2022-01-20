package registry

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
)

// UnknownTypeError is returned by NewObject and NewList when the
// requested object type has not been registered.
type UnknownTypeError struct{ name string }

var _ error = &UnknownTypeError{}

func (u *UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown message type: %q", u.name)
}

func (u *UnknownTypeError) Typename() string {
	return u.name
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		objectTypes:     make(map[string]model.KubernetesObject),
		objectListTypes: make(map[string]model.KubernetesList),
	}
}

var _ TypeRegistry = &typeRegistry{}

type typeRegistry struct {
	objectTypes     map[string]model.KubernetesObject
	objectListTypes map[string]model.KubernetesList
}

func (r *typeRegistry) RegisterObjectType(typ ResourceType, obj model.KubernetesObject) error {
	name := proto.MessageName(typ)
	if previous, ok := r.objectTypes[name]; ok {
		return errors.Errorf("duplicate registration of KubernetesObject type under name %q: previous=%#v new=%#v", name, previous, obj)
	}
	r.objectTypes[name] = obj
	return nil
}

func (r *typeRegistry) RegisterObjectTypeIfAbsent(typ ResourceType, obj model.KubernetesObject) {
	name := proto.MessageName(typ)
	if _, exists := r.objectTypes[name]; exists {
		return
	}
	r.objectTypes[name] = obj
}

func (r *typeRegistry) RegisterListType(typ ResourceType, obj model.KubernetesList) error {
	name := proto.MessageName(typ)
	if previous, ok := r.objectListTypes[name]; ok {
		return errors.Errorf("duplicate registration of KubernetesList type under name %q: previous=%#v new=%#v", name, previous, obj)
	}
	r.objectListTypes[name] = obj
	return nil
}

func (r *typeRegistry) RegisterListTypeIfAbsent(typ ResourceType, obj model.KubernetesList) {
	name := proto.MessageName(typ)
	if _, exists := r.objectListTypes[name]; exists {
		return
	}
	r.objectListTypes[name] = obj
}

func (r *typeRegistry) NewObject(typ ResourceType) (model.KubernetesObject, error) {
	name := proto.MessageName(typ)
	if obj, ok := r.objectTypes[name]; ok {
		return obj.DeepCopyObject().(model.KubernetesObject), nil
	}
	return nil, &UnknownTypeError{name: name}
}

func (r *typeRegistry) NewList(typ ResourceType) (model.KubernetesList, error) {
	name := proto.MessageName(typ)
	if obj, ok := r.objectListTypes[name]; ok {
		return obj.DeepCopyObject().(model.KubernetesList), nil
	}
	return nil, &UnknownTypeError{name: name}
}
