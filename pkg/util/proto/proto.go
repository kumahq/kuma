package proto

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	protov1 "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	json, err := marshaler.MarshalToString(protov1.MessageV1(pb))
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

// we change proto to map, and map to json so we can keep data ordered
// and do not detect changes once resource is returned from k8s
func ToJSONSorted(pb proto.Message) ([]byte, error) {
	jsonData, err := protojson.Marshal(pb)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(jsonData, &resultMap)
	if err != nil {
		return nil, err
	}

	// json.Marshal returns sorted data
	jsonData, err = json.Marshal(resultMap)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func ToJSON(pb proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	if err := marshaler.Marshal(&buf, protov1.MessageV1(pb)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MustMarshalJSONSorted(in proto.Message) []byte {
	content, err := ToJSONSorted(in)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %T: %s", in, err))
	}

	return content
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
	return unmarshaler.Unmarshal(bytes.NewReader(content), protov1.MessageV1(out))
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
