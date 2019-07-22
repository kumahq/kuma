package model

import (
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var _ core_model.ResourceMeta = &ResourceMeta{}

type ResourceMeta struct {
	Mesh      string
	Namespace string
	Name      string
	Version   string
}

func (m *ResourceMeta) GetMesh() string {
	return m.Mesh
}
func (m *ResourceMeta) GetNamespace() string {
	return m.Namespace
}
func (m *ResourceMeta) GetName() string {
	return m.Name
}
func (m *ResourceMeta) GetVersion() string {
	return m.Version
}
