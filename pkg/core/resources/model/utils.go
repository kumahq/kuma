package model

import (
	"encoding/json"
	"path"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

func ToJSON(spec ResourceSpec) ([]byte, error) {
	if msg, ok := spec.(proto.Message); ok {
		return util_proto.ToJSON(msg)
	} else {
		return json.Marshal(spec)
	}
}

func ToMap(spec ResourceSpec) (map[string]any, error) {
	v, err := ToJSON(spec)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
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

// KDSWireSpec is implemented by a spec that a released control plane defines as
// a protobuf message but this one defines as a plain Go struct. ToAny and
// FromAny are the KDS wire codec, and such a peer reads the Any by its type URL
// and then parses protobuf, so the wire form has to stay the message it knows.
// Dropping an implementation does not degrade gracefully: the peer fails the
// whole DeltaDiscoveryResponse rather than the one resource, so its sync stream
// restarts forever and nothing at all reaches it.
type KDSWireSpec interface {
	ToKDSWire() proto.Message
	FromKDSWire(proto.Message) error
}

func ToAny(spec ResourceSpec) (*anypb.Any, error) {
	switch s := spec.(type) {
	case proto.Message:
		return util_proto.MarshalAnyDeterministic(s)
	case KDSWireSpec:
		return util_proto.MarshalAnyDeterministic(s.ToKDSWire())
	default:
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

// FromAny reads a KDSWireSpec from either form: the protobuf message ToAny
// writes, or the type-URL-less JSON a control plane built between the Go struct
// rewrite and this fix sends. Failing that JSON instead would leave such a peer
// restarting its stream with nothing synced.
func FromAny(src *anypb.Any, spec ResourceSpec) error {
	switch s := spec.(type) {
	case proto.Message:
		return util_proto.UnmarshalAnyTo(src, s)
	case KDSWireSpec:
		if src.GetTypeUrl() == "" {
			return json.Unmarshal(src.GetValue(), spec)
		}
		wire := s.ToKDSWire()
		proto.Reset(wire)
		if err := util_proto.UnmarshalAnyTo(src, wire); err != nil {
			return err
		}
		return s.FromKDSWire(wire)
	default:
		return json.Unmarshal(src.GetValue(), spec)
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
