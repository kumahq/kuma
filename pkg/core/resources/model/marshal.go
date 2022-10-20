package model

import (
	"encoding/json"

	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ToJSON(desc ResourceTypeDescriptor, spec ResourceSpec) ([]byte, error) {
	m := &marshal[[]byte]{protoFn: util_proto.ToJSON, fn: json.Marshal}
	return m.ResourceSpec(desc, spec)
}

func ToYAML(desc ResourceTypeDescriptor, spec ResourceSpec) ([]byte, error) {
	m := &marshal[[]byte]{protoFn: util_proto.ToYAML, fn: yaml.Marshal}
	return m.ResourceSpec(desc, spec)
}

func ToAny(desc ResourceTypeDescriptor, spec ResourceSpec) (*anypb.Any, error) {
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
	return m.ResourceSpec(desc, spec)
}

func FromJSON(desc ResourceTypeDescriptor, src []byte, spec ResourceSpec) error {
	u := &unmarshal[[]byte]{protoFn: util_proto.FromJSON, fn: json.Unmarshal}
	return u.ResourceSpec(desc, src, spec)
}

func FromYAML(desc ResourceTypeDescriptor, src []byte, spec ResourceSpec) error {
	u := &unmarshal[[]byte]{protoFn: util_proto.FromYAML, fn: yaml.Unmarshal}
	return u.ResourceSpec(desc, src, spec)
}

func FromAny(desc ResourceTypeDescriptor, src *anypb.Any, spec ResourceSpec) error {
	u := &unmarshal[*anypb.Any]{
		protoFn: util_proto.UnmarshalAnyTo,
		fn: func(a *anypb.Any, dst any) error {
			return json.Unmarshal(a.Value, dst)
		},
	}
	return u.ResourceSpec(desc, src, spec)
}

type marshal[T any] struct {
	protoFn func(proto.Message) (T, error)
	fn      func(any) (T, error)
}

func (m *marshal[T]) ResourceSpec(desc ResourceTypeDescriptor, spec ResourceSpec) (T, error) {
	if desc.IsPluginOriginated {
		return m.fn(spec)
	} else {
		return m.protoFn(spec)
	}
}

type unmarshal[T any] struct {
	protoFn func(T, proto.Message) error
	fn      func(T, any) error
}

func (m *unmarshal[T]) ResourceSpec(desc ResourceTypeDescriptor, src T, spec ResourceSpec) error {
	if desc.IsPluginOriginated {
		return m.fn(src, spec)
	} else {
		return m.protoFn(src, spec)
	}
}
