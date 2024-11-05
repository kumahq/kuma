package model

import (
	"maps"
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var (
	_ core_model.Resource     = &Resource{}
	_ core_model.ResourceMeta = &ResourceMeta{}
)

type Resource struct {
	Meta           core_model.ResourceMeta
	Spec           core_model.ResourceSpec
	Status         core_model.ResourceStatus
	TypeDescriptor core_model.ResourceTypeDescriptor
}

func (r *Resource) GetStatus() core_model.ResourceStatus {
	return r.Status
}

func (r *Resource) SetStatus(status core_model.ResourceStatus) error {
	r.Status = status
	return nil
}

func (r *Resource) SetMeta(meta core_model.ResourceMeta) {
	r.Meta = meta
}

func (r *Resource) SetSpec(spec core_model.ResourceSpec) error {
	r.Spec = spec
	return nil
}

func (r *Resource) GetMeta() core_model.ResourceMeta {
	return r.Meta
}

func (r *Resource) GetSpec() core_model.ResourceSpec {
	return r.Spec
}

func (r *Resource) Descriptor() core_model.ResourceTypeDescriptor {
	return r.TypeDescriptor
}

type ResourceMeta struct {
	Mesh             string
	Name             string
	NameExtensions   core_model.ResourceNameExtensions
	Version          string
	CreationTime     time.Time
	ModificationTime time.Time
	Labels           map[string]string
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

func (m *ResourceMeta) GetLabels() map[string]string {
	return maps.Clone(m.Labels)
}
