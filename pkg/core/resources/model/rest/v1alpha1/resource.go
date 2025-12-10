package v1alpha1

import (
	"encoding/json"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type Resource struct {
	ResourceMeta
	Spec   core_model.ResourceSpec   `json:"spec,omitempty"`
	Status core_model.ResourceStatus `json:"status,omitempty"`
}

var _ json.Marshaler = (*Resource)(nil)

func (r *Resource) MarshalJSON() ([]byte, error) {
	var specJSON json.RawMessage
	if r.Spec != nil {
		b, err := core_model.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
		specJSON = b
	}

	var statusJSON json.RawMessage
	if r.Status != nil {
		b, err := core_model.ToJSON(r.Status)
		if err != nil {
			return nil, err
		}
		statusJSON = b
	}

	var kriStr string
	if id, _ := kri.FromResourceMetaE(r.ResourceMeta, r.Type); !id.IsEmpty() {
		kriStr = id.String()
	}

	// Explicit struct with all fields in desired order
	// Uses standard json.Marshal to avoid manual byte manipulation
	aux := &struct {
		ResourceMeta
		KRI    string          `json:"kri,omitempty"`
		Spec   json.RawMessage `json:"spec,omitempty"`
		Status json.RawMessage `json:"status,omitempty"`
	}{
		ResourceMeta: r.ResourceMeta,
		KRI:          kriStr,
		Spec:         specJSON,
		Status:       statusJSON,
	}

	return json.Marshal(aux)
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
