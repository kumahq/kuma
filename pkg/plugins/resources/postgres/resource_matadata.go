package postgres

import (
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type resourceMetaObject struct {
	Name             string
	Version          string
	Mesh             string
	CreationTime     time.Time
	ModificationTime time.Time
}

var _ core_model.ResourceMeta = &resourceMetaObject{}

func (r *resourceMetaObject) GetName() string {
	return r.Name
}

func (r *resourceMetaObject) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}

func (r *resourceMetaObject) GetVersion() string {
	return r.Version
}

func (r *resourceMetaObject) GetMesh() string {
	return r.Mesh
}

func (r *resourceMetaObject) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r *resourceMetaObject) GetModificationTime() time.Time {
	return r.ModificationTime
}
