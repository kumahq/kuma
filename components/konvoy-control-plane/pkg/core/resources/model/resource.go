package model

import (
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
)

type Resource interface {
	GetType() ResourceType
	GetMeta() ResourceMeta
	SetMeta(ResourceMeta)
	GetSpec() ResourceSpec
}

type ResourceType string

type ResourceMeta interface {
	GetName() string
	GetNamespace() string
	GetVersion() string
}

type ResourceSpec interface {
	// all resources must be defined via Protobuf
	proto.Message
}

type ResourceList interface {
	GetItemType() ResourceType
	NewItem() Resource
	AddItem(Resource) error
}

func ErrorInvalidItemType(expected, actual interface{}) error {
	return fmt.Errorf("Invalid argument type: expected=%q got=%q", reflect.TypeOf(expected), reflect.TypeOf(actual))
}
