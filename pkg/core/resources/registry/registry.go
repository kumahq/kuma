package registry

import (
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/pkg/errors"
	"reflect"
)

type TypeRegistry interface {
	RegisterType(model.Resource) error
	RegisterListType(model.ResourceList) error

	NewObject(model.ResourceType) (model.Resource, error)
	NewList(model.ResourceType) (model.ResourceList, error)

	ObjectTypes() []model.ResourceType
	ListTypes() []model.ResourceType
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

func (t *typeRegistry) ObjectTypes() []model.ResourceType {
	var types []model.ResourceType
	for typ := range t.objectTypes {
		types = append(types, typ)
	}
	return types
}

func (t *typeRegistry) ListTypes() []model.ResourceType {
	var types []model.ResourceType
	for typ := range t.objectListTypes {
		types = append(types, typ)
	}
	return types
}

func (t *typeRegistry) RegisterType(res model.Resource) error {
	newType := reflect.TypeOf(res).Elem()
	if previous, ok := t.objectTypes[res.GetType()]; ok {
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", res.GetType(), previous.String(), newType.String())
	}
	t.objectTypes[res.GetType()] = newType
	return nil
}

func (t *typeRegistry) RegisterListType(resList model.ResourceList) error {
	newType := reflect.TypeOf(resList).Elem()
	if previous, ok := t.objectListTypes[resList.GetItemType()]; ok {
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", resList.GetItemType(), previous.String(), newType.String())
	}
	t.objectListTypes[resList.GetItemType()] = reflect.TypeOf(resList).Elem()
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
