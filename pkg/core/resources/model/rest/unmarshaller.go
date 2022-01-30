package rest

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

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
		return nil, errors.Wrap(err, "invalid meta type")
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
		return nil, errors.Wrapf(err, "invalid %s object %q", meta.Type, meta.Name)
	}
	return res, nil
}
