package proto

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/jsonpb" // nolint: depguard
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/structpb"
	"sigs.k8s.io/yaml"
)

// Note: we continue to use github.com/golang/protobuf/jsonpb because it
// unmarshals types the way we expect in go.
// See https://github.com/golang/protobuf/issues/1374

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func ToYAML(pb proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(protoimpl.X.ProtoMessageV1Of(pb))
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func ToJSON(pb proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(&buf, protoimpl.X.ProtoMessageV1Of(pb)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MustMarshalJSON(in proto.Message) []byte {
	content, err := ToJSON(in)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %T: %s", in, err))
	}

	return content
}

func FromJSON(content []byte, out proto.Message) error {
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(bytes.NewReader(content), protoimpl.X.ProtoMessageV1Of(out))
}

// Converts loosely typed Struct to strongly typed Message
func ToTyped(protoStruct *structpb.Struct, message proto.Message) error {
	if protoStruct == nil {
		return nil
	}
	configBytes, err := ToJSON(protoStruct)
	if err != nil {
		return err
	}
	if err := FromJSON(configBytes, message); err != nil {
		return err
	}
	return nil
}

// Converts loosely typed Struct to strongly typed Message
func ToStruct(message proto.Message) (*structpb.Struct, error) {
	configBytes, err := ToJSON(message)
	if err != nil {
		return nil, err
	}
	str := &structpb.Struct{}
	if err := FromJSON(configBytes, str); err != nil {
		return nil, err
	}
	return str, nil
}

func MustToStruct(message proto.Message) *structpb.Struct {
	str, err := ToStruct(message)
	if err != nil {
		panic(err)
	}
	return str
}
