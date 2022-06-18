package registry

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type TypeRegistry interface {
	RegisterType(model.ResourceTypeDescriptor) error

	NewObject(model.ResourceType) (model.Resource, error)
	NewList(model.ResourceType) (model.ResourceList, error)
	DescriptorFor(resourceType model.ResourceType) (model.ResourceTypeDescriptor, error)

	ObjectTypes(filters ...model.TypeFilter) []model.ResourceType
	ObjectDescriptors(filters ...model.TypeFilter) []model.ResourceTypeDescriptor
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		descriptors: make(map[model.ResourceType]model.ResourceTypeDescriptor),
	}
}

type typeRegistry struct {
	descriptors map[model.ResourceType]model.ResourceTypeDescriptor
}

func (t *typeRegistry) DescriptorFor(resType model.ResourceType) (model.ResourceTypeDescriptor, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return model.ResourceTypeDescriptor{}, fmt.Errorf("invalid resource type %q", resType)
	}
	return typDesc, nil
}

func (t *typeRegistry) ObjectDescriptors(filters ...model.TypeFilter) []model.ResourceTypeDescriptor {
	var descriptors []model.ResourceTypeDescriptor
	for _, typ := range t.descriptors {
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
	for _, typ := range t.descriptors {
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
	if res.Resource.GetSpec() == nil {
		return errors.New("spec in the object cannot be nil")
	}
	if previous, ok := t.descriptors[res.Name]; ok {
		return fmt.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", res.Name, previous, reflect.TypeOf(res.Resource).Elem().String())
	}
	t.descriptors[res.Name] = res
	return nil
}

func (t *typeRegistry) NewObject(resType model.ResourceType) (model.Resource, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return nil, fmt.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewObject(), nil
}

func (t *typeRegistry) NewList(resType model.ResourceType) (model.ResourceList, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return nil, fmt.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewList(), nil
}
