package registry

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type TypeRegistry interface {
	RegisterType(model.ResourceTypeDescriptor) error

	NewObject(model.ResourceType) (model.Resource, error)
	NewList(model.ResourceType) (model.ResourceList, error)
	Descriptor(resourceType model.ResourceType) (model.ResourceTypeDescriptor, error)

	ObjectTypes(filters ...model.TypeFilter) []model.ResourceType
	ObjectDesc(filters ...model.TypeFilter) []model.ResourceTypeDescriptor
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		objectSpecs: make(map[model.ResourceType]model.ResourceTypeDescriptor),
	}
}

type typeRegistry struct {
	objectSpecs map[model.ResourceType]model.ResourceTypeDescriptor
}

func (t *typeRegistry) Descriptor(resType model.ResourceType) (model.ResourceTypeDescriptor, error) {
	typDesc, ok := t.objectSpecs[resType]
	if !ok {
		return model.ResourceTypeDescriptor{}, errors.Errorf("invalid resource type %q", resType)
	}
	return typDesc, nil
}

func (t *typeRegistry) ObjectDesc(filters ...model.TypeFilter) []model.ResourceTypeDescriptor {
	var descriptors []model.ResourceTypeDescriptor
	for _, typ := range t.objectSpecs {
		match := true
		for _, f := range filters {
			match = match && f.Apply(typ)
		}
		if match {
			descriptors = append(descriptors, typ)
		}
	}
	return descriptors
}

func (t *typeRegistry) ObjectTypes(filters ...model.TypeFilter) []model.ResourceType {
	var types []model.ResourceType
	for _, typ := range t.objectSpecs {
		match := true
		for _, f := range filters {
			match = match && f.Apply(typ)
		}
		if match {
			types = append(types, typ.Name)
		}
	}
	return types
}

func (t *typeRegistry) RegisterType(res model.ResourceTypeDescriptor) error {
	newType := reflect.TypeOf(res.Resource).Elem()
	if previous, ok := t.objectSpecs[res.Name]; ok {
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", res.Name, previous, newType.String())
	}
	if res.Resource.GetSpec() == nil {
		return errors.New("spec in the object cannot be nil")
	}
	res.ObjectType = newType
	if res.ResourceList != nil {
		res.ListType = reflect.TypeOf(res.ResourceList).Elem()
	}
	t.objectSpecs[res.Name] = res
	return nil
}

func (t *typeRegistry) NewObject(resType model.ResourceType) (model.Resource, error) {
	typDesc, ok := t.objectSpecs[resType]
	if !ok {
		return nil, errors.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewObject(), nil
}

func (t *typeRegistry) NewList(resType model.ResourceType) (model.ResourceList, error) {
	typDesc, ok := t.objectSpecs[resType]
	if !ok {
		return nil, errors.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewList(), nil
}
