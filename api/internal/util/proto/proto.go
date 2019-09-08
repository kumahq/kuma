package proto

import (
	"bytes"

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

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func FromJSON(content []byte, out proto.Message) error {
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(bytes.NewReader(content), out)
}
