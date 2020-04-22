package proto

import (
	"bytes"
	"encoding/json"

	pstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
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

func FromJSON(content []byte, out proto.Message) error {
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(bytes.NewReader(content), out)
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
func ToTyped(protoStruct *pstruct.Struct, message proto.Message) error {
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
