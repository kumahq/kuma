package registry

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type TypeRegistry interface {
	RegisterType(model.ResourceTypeDescriptor) error

	NewObject(model.ResourceType) (model.Resource, error)
	NewList(model.ResourceType) (model.ResourceList, error)

	MustNewObject(model.ResourceType) model.Resource
	MustNewList(model.ResourceType) model.ResourceList

	DescriptorFor(resourceType model.ResourceType) (model.ResourceTypeDescriptor, error)

	ObjectTypes(filters ...model.TypeFilter) []model.ResourceType
	ObjectDescriptors(filters ...model.TypeFilter) []model.ResourceTypeDescriptor
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		descriptors: make(map[model.ResourceType]model.ResourceTypeDescriptor),
	}
}

type InvalidResourceTypeError struct {
	ResType model.ResourceType
}

func (e *InvalidResourceTypeError) Error() string {
	return fmt.Sprintf("invalid resource type %q", e.ResType)
}

func (e *InvalidResourceTypeError) Is(target error) bool {
	t, ok := target.(*InvalidResourceTypeError)
	if !ok {
		return false
	}
	return t.ResType == e.ResType || t.ResType == ""
}

type typeRegistry struct {
	descriptors map[model.ResourceType]model.ResourceTypeDescriptor
}

func (t *typeRegistry) DescriptorFor(resType model.ResourceType) (model.ResourceTypeDescriptor, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return model.ResourceTypeDescriptor{}, &InvalidResourceTypeError{ResType: resType}
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
	// check no duplicated short name
	if res.ShortName != "" {
		for _, descriptor := range t.descriptors {
			if descriptor.ShortName == res.ShortName {
				return errors.Errorf("duplicate registration of ResourceType under short name %q: previous=%#v new=%#v", res.ShortName, descriptor.Name, res.Name)
			}
		}
	}
	t.descriptors[res.Name] = res
	return nil
}

func (t *typeRegistry) NewObject(resType model.ResourceType) (model.Resource, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return nil, errors.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewObject(), nil
}

func (t *typeRegistry) NewList(resType model.ResourceType) (model.ResourceList, error) {
	typDesc, ok := t.descriptors[resType]
	if !ok {
		return nil, errors.Errorf("invalid resource type %q", resType)
	}
	return typDesc.NewList(), nil
}

// MustNewObject implements TypeRegistry.
func (t *typeRegistry) MustNewObject(resType model.ResourceType) model.Resource {
	res, err := t.NewObject(resType)
	if err != nil {
		panic(err)
	}
	return res
}

// MustNewList implements TypeRegistry.
func (t *typeRegistry) MustNewList(resType model.ResourceType) model.ResourceList {
	resList, err := t.NewList(resType)
	if err != nil {
		panic(err)
	}
	return resList
}
