package model

import (
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ core_model.ResourceMeta = &ResourceMeta{}

type ResourceMeta struct {
	Mesh             string
	Name             string
	NameExtensions   core_model.ResourceNameExtensions
	Version          string
	CreationTime     time.Time
	ModificationTime time.Time
}

func (m *ResourceMeta) GetMesh() string {
	return m.Mesh
}

func (m *ResourceMeta) GetName() string {
	return m.Name
}

func (m *ResourceMeta) GetNameExtensions() core_model.ResourceNameExtensions {
	return m.NameExtensions
}

func (m *ResourceMeta) GetVersion() string {
	return m.Version
}

func (m *ResourceMeta) GetCreationTime() time.Time {
	return m.CreationTime
}

func (m *ResourceMeta) GetModificationTime() time.Time {
	return m.ModificationTime
}
