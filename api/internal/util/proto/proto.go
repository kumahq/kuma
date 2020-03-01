package proto

import (
	"bytes"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func ToYAML(pb proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(pb)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
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
