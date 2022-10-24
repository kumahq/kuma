package model

import (
	"encoding/json"
	"path"
	"reflect"

	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ToJSON(desc ResourceTypeDescriptor, spec ResourceSpec) ([]byte, error) {
	if !desc.IsPluginOriginated {
		return util_proto.ToJSON(spec.(proto.Message))
	} else {
		return json.Marshal(spec)
	}
}

func ToYAML(desc ResourceTypeDescriptor, spec ResourceSpec) ([]byte, error) {
	if !desc.IsPluginOriginated {
		return util_proto.ToYAML(spec.(proto.Message))
	} else {
		return yaml.Marshal(spec)
	}
}

func ToAny(desc ResourceTypeDescriptor, spec ResourceSpec) (*anypb.Any, error) {
	if !desc.IsPluginOriginated {
		return util_proto.MarshalAnyDeterministic(spec.(proto.Message))
	} else {
		bytes, err := json.Marshal(spec)
		if err != nil {
			return nil, err
		}
		return &anypb.Any{
			Value: bytes,
		}, nil
	}
}

func FromJSON(desc ResourceTypeDescriptor, src []byte, spec ResourceSpec) error {
	if !desc.IsPluginOriginated {
		return util_proto.FromJSON(src, spec.(proto.Message))
	} else {
		return json.Unmarshal(src, spec)
	}
}

func FromYAML(desc ResourceTypeDescriptor, src []byte, spec ResourceSpec) error {
	if !desc.IsPluginOriginated {
		return util_proto.FromYAML(src, spec.(proto.Message))
	} else {
		return yaml.Unmarshal(src, spec)
	}
}

func FromAny(desc ResourceTypeDescriptor, src *anypb.Any, spec ResourceSpec) error {
	if !desc.IsPluginOriginated {
		return util_proto.UnmarshalAnyTo(src, spec.(proto.Message))
	} else {
		return json.Unmarshal(src.Value, spec)
	}
}

func FullName(spec ResourceSpec) string {
	specType := reflect.TypeOf(spec).Elem()
	return path.Join(specType.PkgPath(), specType.Name())
}

func Equal(x, y ResourceSpec) bool {
	xMsg, xOk := x.(proto.Message)
	yMsg, yOk := y.(proto.Message)
	if xOk != yOk {
		return false
	}

	if xOk {
		return proto.Equal(xMsg, yMsg)
	} else {
		return reflect.DeepEqual(x, y)
	}
}

func IsEmpty(spec ResourceSpec) bool {
	if msg, ok := spec.(proto.Message); ok {
		return proto.Size(msg) == 0
	} else {
		return reflect.ValueOf(spec).Elem().IsZero()
	}
}
