package proto

import (
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
