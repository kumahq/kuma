package model

import (
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	"time"
)

var _ core_model.ResourceMeta = &ResourceMeta{}

type ResourceMeta struct {
	Mesh             string
	Name             string
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
func (m *ResourceMeta) GetVersion() string {
	return m.Version
}
func (m *ResourceMeta) GetCreationTime() time.Time {
	return m.CreationTime
}
func (m *ResourceMeta) GetModificationTime() time.Time {
	return m.ModificationTime
}
