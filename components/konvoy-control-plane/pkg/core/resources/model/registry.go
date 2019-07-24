package model

import (
	"reflect"

	"github.com/pkg/errors"
)

type ResourceRegistry interface {
	NewResource(ResourceType) (Resource, error)
}

var _ ResourceRegistry = &SimpleResourceRegistry{}

type SimpleResourceRegistry struct {
	ResourceTypes map[ResourceType]Resource
}

func (reg *SimpleResourceRegistry) NewResource(typ ResourceType) (Resource, error) {
	r, ok := reg.ResourceTypes[typ]
	if !ok {
		return nil, errors.Errorf("unknown resource type: %q", typ)
	}
	prototype := reflect.ValueOf(r)
	copy := reflect.New(prototype.Type().Elem())
	return copy.Interface().(Resource), nil
}
