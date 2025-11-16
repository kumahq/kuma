package v1alpha1

import (
	"encoding/json"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type Resource struct {
	ResourceMeta
	Spec   core_model.ResourceSpec   `json:"spec,omitempty"`
	Status core_model.ResourceStatus `json:"status,omitempty"`
}

var _ json.Marshaler = (*Resource)(nil)

func (r *Resource) MarshalJSON() ([]byte, error) {
	metaBytes, err := json.Marshal(r.ResourceMeta) // includes "kri" from ResourceMeta.MarshalJSON
	if err != nil {
		return nil, err
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(metaBytes, &out); err != nil {
		return nil, err
	}

	if r.Spec != nil {
		b, err := core_model.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
		out["spec"] = b
	}

	if r.Status != nil {
		b, err := core_model.ToJSON(r.Status)
		if err != nil {
			return nil, err
		}
		out["status"] = b
	}

	return json.Marshal(out)
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

func (r *Resource) GetStatus() core_model.ResourceStatus {
	if r == nil {
		return nil
	}
	return r.Status
}
