package registry

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type TypeRegistry interface {
	RegisterType(model.ResourceTypeDescriptor) error

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

func Must(resType model.ResourceTypeDescriptor, err error) model.ResourceTypeDescriptor {
	if err != nil {
		panic(err)
	}
	return resType
}

func (t *typeRegistry) DescriptorFor(resType model.ResourceType) (model.ResourceTypeDescriptor, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return model.ResourceTypeDescriptor{}, errors.Errorf("invalid resource type %q", resType)
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
		return errors.Errorf("duplicate registration of ResourceType under name %q: previous=%#v new=%#v", res.Name, previous, reflect.TypeOf(res.Resource).Elem().String())
	}
	t.descriptors[res.Name] = res
	return nil
}
