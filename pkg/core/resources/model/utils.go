package model

import (
	"encoding/json"
	"path"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ToJSON(spec ResourceSpec) ([]byte, error) {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.ToJSON(msg)
	} else {
		return json.Marshal(spec)
	}
}

func ToMap(spec ResourceSpec) (map[string]interface{}, error) {
	v, err := ToJSON(spec)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	if err := json.Unmarshal(v, &result); err != nil {
		return result, err
	}
	return result, nil
}

func ToYAML(spec ResourceSpec) ([]byte, error) {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.ToYAML(msg)
	} else {
		return yaml.Marshal(spec)
	}
}

func ToAny(spec ResourceSpec) (*anypb.Any, error) {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.MarshalAnyDeterministic(msg)
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

func FromJSON(src []byte, spec ResourceSpec) error {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.FromJSON(src, msg)
	} else {
		return json.Unmarshal(src, spec)
	}
}

func FromYAML(src []byte, spec ResourceSpec) error {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.FromYAML(src, msg)
	} else {
		return yaml.Unmarshal(src, spec)
	}
}

func FromAny(src *anypb.Any, spec ResourceSpec) error {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.UnmarshalAnyTo(src, msg)
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
