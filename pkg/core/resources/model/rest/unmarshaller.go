package rest

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

var YAML = &unmarshaler{unmarshalFn: func(bytes []byte, i interface{}) error {
	return yaml.Unmarshal(bytes, i)
}}
var JSON = &unmarshaler{unmarshalFn: json.Unmarshal}

type unmarshaler struct {
	unmarshalFn func([]byte, interface{}) error
}

func (u *unmarshaler) UnmarshalCore(bytes []byte) (core_model.Resource, error) {
	restResource, err := u.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}
	coreRes, err := To.Core(restResource)
	if err != nil {
		return nil, err
	}
	return coreRes, nil
}

func (u *unmarshaler) UnmarshalToCore(bytes []byte, res core_model.Resource) error {
	restResource, err := u.Unmarshal(bytes)
	if err != nil {
		return err
	}
	coreRes, err := To.Core(restResource)
	if err != nil {
		return err
	}
	res.SetMeta(coreRes.GetMeta())
	if err := res.SetSpec(coreRes.GetSpec()); err != nil {
		return err
	}
	return nil
}

func (u *unmarshaler) Unmarshal(bytes []byte) (Resource, error) {
	meta := v1alpha1.ResourceMeta{}
	if err := u.unmarshalFn(bytes, &meta); err != nil {
		return nil, errors.Wrap(err, "invalid meta type")
	}

	desc, err := registry.Global().DescriptorFor(core_model.ResourceType(meta.Type))
	if err != nil {
		return nil, err
	}

	resource := desc.NewObject()
	restResource := From.Resource(resource)

	if desc.IsPluginOriginated {
		// desc.Schema is set only for new plugin originated policies
		if err := u.rawSchemaValidation(bytes, desc.Schema); err != nil {
			return nil, err
		}
	}

	if err := u.unmarshalFn(bytes, restResource); err != nil {
		return nil, errors.Wrapf(err, "invalid %s object %q", meta.Type, meta.Name)
	}

	if err := core_model.Validate(resource); err != nil {
		return nil, err
	}

	return restResource, nil
}

func (u *unmarshaler) UnmarshalListToCore(b []byte, rs core_model.ResourceList) error {
	rsr := &ResourceListReceiver{
		NewResource: rs.Descriptor().NewObject,
	}
	if err := u.unmarshalFn(b, rsr); err != nil {
		return err
	}
	for _, ri := range rsr.ResourceList.Items {
		r := rsr.NewResource()
		if err := r.SetSpec(ri.GetSpec()); err != nil {
			return err
		}
		r.SetMeta(ri.GetMeta())
		_ = rs.AddItem(r)
	}
	if rsr.Next != nil {
		uri, err := url.ParseRequestURI(*rsr.Next)
		if err != nil {
			return errors.Wrap(err, "invalid next URL from the server")
		}
		offset := uri.Query().Get("offset")
		// we do not preserve here the size of the page, but since it is used in kumactl
		// user will rerun command with the page size of his choice
		if offset != "" {
			rs.GetPagination().SetNextOffset(offset)
		}
	}
	rs.GetPagination().SetTotal(rsr.ResourceList.Total)
	return nil
}
