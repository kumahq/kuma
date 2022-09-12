package v1alpha1

import (
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceMeta struct {
	Type             string    `json:"type"`
	Mesh             string    `json:"mesh,omitempty"`
	Name             string    `json:"name"`
	CreationTime     time.Time `json:"creationTime"`
	ModificationTime time.Time `json:"modificationTime"`
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
