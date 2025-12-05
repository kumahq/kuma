package v1alpha1

import (
	"bytes"
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

	var specBytes []byte
	if r.Spec != nil {
		specBytes, err = core_model.ToJSON(r.Spec)
		if err != nil {
			return nil, err
		}
	}

	var statusBytes []byte
	if r.Status != nil {
		statusBytes, err = core_model.ToJSON(r.Status)
		if err != nil {
			return nil, err
		}
	}

	// Manually concatenate JSON bytes to preserve field order
	// metaBytes is like: {"type":"Mesh",...}
	// We need to build: {"type":"Mesh",...,"spec":{...},"status":{...}}

	result := metaBytes[:len(metaBytes)-1] // Remove closing }

	if len(specBytes) > 0 && !bytes.Equal(specBytes, []byte("{}")) {
		result = append(result, []byte(`,"spec":`)...)
		result = append(result, specBytes...)
	}

	if len(statusBytes) > 0 && !bytes.Equal(statusBytes, []byte("{}")) {
		result = append(result, []byte(`,"status":`)...)
		result = append(result, statusBytes...)
	}

	result = append(result, '}') // Add closing }

	return result, nil
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
