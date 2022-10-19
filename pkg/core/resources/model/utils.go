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

func ToJSON(spec ResourceSpec) ([]byte, error) {
	m := &marshal[[]byte]{protoFn: util_proto.ToJSON, fn: json.Marshal}
	return m.ResourceSpec(spec)
}

func ToYAML(spec ResourceSpec) ([]byte, error) {
	m := &marshal[[]byte]{protoFn: util_proto.ToYAML, fn: yaml.Marshal}
	return m.ResourceSpec(spec)
}

func ToAny(spec ResourceSpec) (*anypb.Any, error) {
	m := &marshal[*anypb.Any]{
		protoFn: util_proto.MarshalAnyDeterministic,
		fn: func(a any) (*anypb.Any, error) {
			bytes, err := json.Marshal(a)
			if err != nil {
				return nil, err
			}
			return &anypb.Any{
				Value: bytes,
			}, nil
		},
	}
	return m.ResourceSpec(spec)
}

func FromJSON(src []byte, spec ResourceSpec) error {
	u := &unmarshal[[]byte]{protoFn: util_proto.FromJSON, fn: json.Unmarshal}
	return u.ResourceSpec(src, spec)
}

func FromYAML(src []byte, spec ResourceSpec) error {
	u := &unmarshal[[]byte]{protoFn: util_proto.FromYAML, fn: yaml.Unmarshal}
	return u.ResourceSpec(src, spec)
}

func FromAny(src *anypb.Any, spec ResourceSpec) error {
	u := &unmarshal[*anypb.Any]{
		protoFn: util_proto.UnmarshalAnyTo,
		fn: func(a *anypb.Any, dst any) error {
			return json.Unmarshal(a.Value, dst)
		},
	}
	return u.ResourceSpec(src, spec)
}

type marshal[T any] struct {
	protoFn func(proto.Message) (T, error)
	fn      func(any) (T, error)
}

func (m *marshal[T]) ResourceSpec(spec ResourceSpec) (T, error) {
	if msg, ok := spec.(proto.Message); ok {
		return m.protoFn(msg)
	} else {
		return m.fn(spec)
	}
}

type unmarshal[T any] struct {
	protoFn func(T, proto.Message) error
	fn      func(T, any) error
}

func (m *unmarshal[T]) ResourceSpec(src T, spec ResourceSpec) error {
	if msg, ok := spec.(proto.Message); ok {
		return m.protoFn(src, msg)
	} else {
		return m.fn(src, spec)
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
