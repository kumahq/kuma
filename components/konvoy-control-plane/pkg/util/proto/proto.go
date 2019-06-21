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

func FromMap(in map[string]interface{}, out proto.Message) error {
	buf, err := json.Marshal(in)
	if err != nil {
		return err
	}
	unmarshaler := &jsonpb.Unmarshaler{}
	return unmarshaler.Unmarshal(bytes.NewReader(buf), out)
}
