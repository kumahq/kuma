package unversioned

import (
	"encoding/json"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
)

type Resource struct {
	Meta v1alpha1.ResourceMeta
	Spec core_model.ResourceSpec
}

func (r *Resource) GetMeta() v1alpha1.ResourceMeta {
	if r == nil {
		return v1alpha1.ResourceMeta{}
	}
	return r.Meta
}

func (r *Resource) GetSpec() core_model.ResourceSpec {
	if r == nil {
		return nil
	}
	return r.Spec
}

func (r *Resource) GetStatus() core_model.ResourceStatus {
	return nil // there is no status in old resources
}

var (
	_ json.Marshaler   = &Resource{}
	_ json.Unmarshaler = &Resource{}
)

func (r *Resource) MarshalJSON() ([]byte, error) {
	var specBytes []byte
	if r.Spec != nil {
		bytes, err := core_model.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
		specBytes = bytes
	}

	metaJSON, err := json.Marshal(r.Meta)
	if err != nil {
		return nil, err
	}

	if len(specBytes) == 0 || string(specBytes) == "{}" { // spec is nil or empty
		return metaJSON, nil
	} else {
		// remove the } of meta JSON, { of spec JSON and join it by ,
		return append(append(metaJSON[:len(metaJSON)-1], byte(',')), specBytes[1:]...), nil
	}
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Meta); err != nil {
		return err
	}
	if r.Spec == nil {
		newR, err := registry.Global().NewObject(core_model.ResourceType(r.Meta.Type))
		if err != nil {
			return err
		}
		r.Spec = newR.GetSpec()
	}
	// The inlined form puts meta and spec fields in one object, so the whole document
	// is handed to the spec decoder and the meta keys are ignored as unknown fields.
	// core_model.FromJSON picks the protobuf or the plain JSON decoder, which is what
	// lets a resource whose spec is a Go struct keep this representation.
	if err := core_model.FromJSON(data, r.Spec); err != nil {
		return err
	}
	return nil
}

func (r *Resource) ToCore() (core_model.Resource, error) {
	resource, err := registry.Global().NewObject(core_model.ResourceType(r.Meta.Type))
	if err != nil {
		return nil, err
	}
	resource.SetMeta(&r.Meta)
	if err := resource.SetSpec(r.Spec); err != nil {
		return nil, err
	}
	return resource, nil
}
