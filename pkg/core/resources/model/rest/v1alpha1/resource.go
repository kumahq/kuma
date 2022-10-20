package v1alpha1

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
