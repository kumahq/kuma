package model

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

const (
	DefaultMesh      = "default"
	DefaultNamespace = "default"
)

type ResourceKey struct {
	Mesh      string
	Namespace string
	Name      string
}

type Resource interface {
	GetType() ResourceType
	GetMeta() ResourceMeta
	SetMeta(ResourceMeta)
	GetSpec() ResourceSpec
	SetSpec(ResourceSpec) error
}

type ResourceType string

type ResourceMeta interface {
	GetName() string
	GetNamespace() string
	GetVersion() string
	GetMesh() string
}

func MetaToResourceKey(meta ResourceMeta) ResourceKey {
	return ResourceKey{
		Mesh:      meta.GetMesh(),
		Namespace: meta.GetNamespace(),
		Name:      meta.GetName(),
	}
}

type ResourceSpec interface {
	// all resources must be defined via Protobuf
	proto.Message
}

type ResourceList interface {
	GetItemType() ResourceType
	GetItems() []Resource
	NewItem() Resource
	AddItem(Resource) error
}

func ErrorInvalidItemType(expected, actual interface{}) error {
	return fmt.Errorf("Invalid argument type: expected=%q got=%q", reflect.TypeOf(expected), reflect.TypeOf(actual))
}
