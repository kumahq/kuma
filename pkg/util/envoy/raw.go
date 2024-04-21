package envoy

import (
	"errors"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func ResourceFromYaml(resYaml string) (proto.Message, error) {
	json, err := yaml.YAMLToJSON([]byte(resYaml))
	if err != nil {
		json = []byte(resYaml)
	}

	var anything anypb.Any
	if err := util_proto.FromJSON(json, &anything); err != nil {
		return nil, err
	}
	msg, err := anything.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	p, ok := msg.(envoy_types.Resource)
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
