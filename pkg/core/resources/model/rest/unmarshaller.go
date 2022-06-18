package rest

import (
	"fmt"

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
	meta := ResourceMeta{}
	if err := yaml.Unmarshal(bytes, &meta); err != nil {
		return nil, fmt.Errorf("invalid meta type: %w", err)
	}
	resource, err := registry.Global().NewObject(model.ResourceType(meta.Type))
	if err != nil {
		return nil, err
	}
	res := &Resource{
		Meta: meta,
		Spec: resource.GetSpec(),
	}
	if err := proto.FromYAML(bytes, res.Spec); err != nil {
		return nil, fmt.Errorf("invalid %s object %q: %w", meta.Type, meta.Name, err)
	}
	return res, nil
}
