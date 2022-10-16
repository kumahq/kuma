package model

import (
	"encoding/json"

	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var ToJSON = &marshal[[]byte]{protoFn: util_proto.ToJSON, fn: json.Marshal}
var ToYAML = &marshal[[]byte]{protoFn: util_proto.ToYAML, fn: yaml.Marshal}
var ToAny = &marshal[*anypb.Any]{
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

var FromJSON = &unmarshal[[]byte]{protoFn: util_proto.FromJSON, fn: json.Unmarshal}
var FromYAML = &unmarshal[[]byte]{protoFn: util_proto.FromYAML, fn: yaml.Unmarshal}
var FromAny = &unmarshal[*anypb.Any]{
	protoFn: util_proto.UnmarshalAnyTo,
	fn: func(a *anypb.Any, dst any) error {
		return json.Unmarshal(a.Value, dst)
	},
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
