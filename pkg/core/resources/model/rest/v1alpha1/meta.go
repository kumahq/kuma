package v1alpha1

import (
	"maps"
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceMeta struct {
	Type             string            `json:"type"`
	Mesh             string            `json:"mesh,omitempty"`
	Name             string            `json:"name"`
	CreationTime     time.Time         `json:"creationTime"`
	ModificationTime time.Time         `json:"modificationTime"`
	Labels           map[string]string `json:"labels,omitempty"`
}

var _ core_model.ResourceMeta = ResourceMeta{}

func (r ResourceMeta) GetName() string {
	return r.Name
}

func (r ResourceMeta) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}

func (r ResourceMeta) GetVersion() string {
	return ""
}

func (r ResourceMeta) GetMesh() string {
	return r.Mesh
}

func (r ResourceMeta) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r ResourceMeta) GetModificationTime() time.Time {
	return r.ModificationTime
}

func (r ResourceMeta) GetLabels() map[string]string {
	return maps.Clone(r.Labels)
}
