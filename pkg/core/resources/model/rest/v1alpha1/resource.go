package v1alpha1

import (
	"encoding/json"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type Resource struct {
	ResourceMeta
	Spec core_model.ResourceSpec `json:"spec,omitempty"`
}

func (r *Resource) GetMeta() ResourceMeta {
	if r == nil {
		return ResourceMeta{}
	}
	return r.ResourceMeta
}

func (r *Resource) GetSpec() core_model.ResourceSpec {
	if r == nil {
		return nil
	}
	return r.Spec
}

var _ json.Marshaler = &Resource{}
var _ json.Unmarshaler = &Resource{}

func (r *Resource) MarshalJSON() ([]byte, error) {
	var specBytes []byte
	var err error
	if r.Spec != nil {
		specBytes, err = proto.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		ResourceMeta
		Spec json.RawMessage `json:"spec,omitempty"`
	}{
		ResourceMeta: r.ResourceMeta,
		Spec:         specBytes,
	})
}

func (r *Resource) UnmarshalJSON(bytes []byte) error {
	obj := &struct {
		ResourceMeta
		Spec json.RawMessage
	}{}
	if err := json.Unmarshal(bytes, obj); err != nil {
		return err
	}

	r.ResourceMeta = obj.ResourceMeta

	if r.Spec == nil {
		newR, err := registry.Global().NewObject(core_model.ResourceType(r.Type))
		if err != nil {
			return err
		}
		r.Spec = newR.GetSpec()
	}

	if err := proto.FromJSON(obj.Spec, r.Spec); err != nil {
		return err
	}
	return nil
}
