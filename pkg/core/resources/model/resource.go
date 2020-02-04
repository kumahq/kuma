package model

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

const (
	DefaultMesh = "default"
)

type ResourceKey struct {
	Mesh string
	Name string
}

type Resource interface {
	GetType() ResourceType
	GetMeta() ResourceMeta
	SetMeta(ResourceMeta)
	GetSpec() ResourceSpec
	SetSpec(ResourceSpec) error
	Validate() error
}

type ResourceType string

type ResourceMeta interface {
	GetName() string
	GetVersion() string
	GetMesh() string
	GetCreationTime() time.Time
	GetModificationTime() time.Time
}

func MetaToResourceKey(meta ResourceMeta) ResourceKey {
	if meta == nil {
		return ResourceKey{}
	}
	return ResourceKey{
		Mesh: meta.GetMesh(),
		Name: meta.GetName(),
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
