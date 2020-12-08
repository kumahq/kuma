package envoy

import (
	"bytes"
	"errors"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func ResourceFromYaml(resYaml string) (proto.Message, error) {
	json, err := yaml.YAMLToJSON([]byte(resYaml))
	if err != nil {
		json = []byte(resYaml)
	}

	var anything any.Any
	if err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewReader(json), &anything); err != nil {
		return nil, err
	}
	var dyn ptypes.DynamicAny
	if err := ptypes.UnmarshalAny(&anything, &dyn); err != nil {
		return nil, err
	}
	p, ok := dyn.Message.(envoy_types.Resource)
	if !ok {
		return nil, errors.New("xDS resource doesn't implement all required interfaces")
	}
	if v, ok := p.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}
	return p, nil
}
