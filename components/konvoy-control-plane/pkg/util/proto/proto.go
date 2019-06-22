package proto

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

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
	unmarshaler := &jsonpb.Unmarshaler{}
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
