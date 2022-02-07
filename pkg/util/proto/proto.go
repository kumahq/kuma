package proto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func ToYAML(pb proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(pb)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func ToJSON(pb proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(&buf, pb); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ToJSONIndent(pb proto.Message, indent string) ([]byte, error) {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{
		Indent: indent,
	}
	if err := marshaler.Marshal(&buf, pb); err != nil {
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
	return unmarshaler.Unmarshal(bytes.NewReader(content), out)
}

func MustUnmarshalJSON(content []byte, out proto.Message) proto.Message {
	if err := FromJSON(content, out); err != nil {
		panic(fmt.Sprintf("failed to unmarshal %T: %s", out, err))
	}

	return out
}

func ToMap(pb proto.Message) (map[string]interface{}, error) {
	content, err := ToJSON(pb)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]interface{})
	if err := json.Unmarshal(content, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func FromMap(in map[string]interface{}, out proto.Message) error {
	content, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return FromJSON(content, out)
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

func IsEmpty(message proto.Message) bool {
	return proto.Size(message) == 0
}
