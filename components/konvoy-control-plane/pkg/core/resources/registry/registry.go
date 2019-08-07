package registry

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/pkg/errors"
	"reflect"
)

type TypeRegistry interface {
	RegisterType(model.ResourceType, model.Resource) error
	RegisterListType(model.ResourceType, model.ResourceList) error

	NewObject(model.ResourceType) (model.Resource, error)
	NewList(model.ResourceType) (model.ResourceList, error)
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		objectTypes:     make(map[model.ResourceType]reflect.Type),
		objectListTypes: make(map[model.ResourceType]reflect.Type),
	}
}

type typeRegistry struct {
	objectTypes     map[model.ResourceType]reflect.Type
	objectListTypes map[model.ResourceType]reflect.Type
}

func (t *typeRegistry) RegisterType(resType model.ResourceType, res model.Resource) error {
	newType := reflect.TypeOf(res).Elem()
	if previous, ok := t.objectTypes[resType]; ok {
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", resType, previous.String(), newType.String())
	}
	t.objectTypes[resType] = newType
	return nil
}

func (t *typeRegistry) RegisterListType(resType model.ResourceType, resList model.ResourceList) error {
	newType := reflect.TypeOf(resList).Elem()
	if previous, ok := t.objectListTypes[resType]; ok {
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", resType, previous.String(), newType.String())
	}
	t.objectListTypes[resType] = reflect.TypeOf(resList).Elem()
	return nil
}

func (t *typeRegistry) NewObject(resType model.ResourceType) (model.Resource, error) {
	typ, ok := t.objectTypes[resType]
	if ok != true {
		return nil, errors.New("invalid type of resource type")
	}
	return reflect.New(typ).Interface().(model.Resource), nil
}

func (t *typeRegistry) NewList(resType model.ResourceType) (model.ResourceList, error) {
	typ, ok := t.objectListTypes[resType]
	if ok != true {
		return nil, errors.New("invalid type of resource type")
	}
	return reflect.New(typ).Interface().(model.ResourceList), nil
}
