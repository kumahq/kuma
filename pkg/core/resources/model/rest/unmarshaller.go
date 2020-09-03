package rest

import (
	"github.com/ghodss/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/util/proto"
)

func UnmarshallToCore(bytes []byte) (model.Resource, error) {
	res, err := Unmarshall(bytes)
	if err != nil {
		return nil, err
	}
	return res.ToCore()
}

func Unmarshall(bytes []byte) (*Resource, error) {
	resMeta := ResourceMeta{}
	if err := yaml.Unmarshal(bytes, &resMeta); err != nil {
		return nil, err
	}
	resource, err := registry.Global().NewObject(model.ResourceType(resMeta.Type))
	if err != nil {
		return nil, err
	}
	res := &Resource{
		Meta: resMeta,
		Spec: resource.GetSpec(),
	}
	if err := proto.FromYAML(bytes, res.Spec); err != nil {
		return nil, err
	}
	return res, nil
}
