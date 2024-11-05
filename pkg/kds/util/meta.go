package util

import (
	"time"

	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// KDS ResourceMeta only contains name and mesh.
// The rest is managed by the receiver of resources anyways. See ResourceSyncer#Sync
type resourceMeta struct {
	name   string
	mesh   string
	labels map[string]string
}

type CloneResourceMetaOpt func(*resourceMeta)

func WithName(name string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		if m.labels[mesh_proto.DisplayName] == "" {
			m.labels[mesh_proto.DisplayName] = m.name
		}
		m.name = name
	}
}

func WithLabel(key, value string) CloneResourceMetaOpt {
	return func(m *resourceMeta) {
		m.labels[key] = value
	}
}

func CloneResourceMeta(m model.ResourceMeta, fs ...CloneResourceMetaOpt) model.ResourceMeta {
	labels := m.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	meta := &resourceMeta{
		name:   m.GetName(),
		mesh:   m.GetMesh(),
		labels: labels,
	}
	for _, f := range fs {
		f(meta)
	}
	if len(meta.labels) == 0 {
		meta.labels = nil
	}
	return meta
}

func kumaResourceMetaToResourceMeta(meta *mesh_proto.KumaResource_Meta) model.ResourceMeta {
	return &resourceMeta{
		name:   meta.Name,
		mesh:   meta.Mesh,
		labels: meta.GetLabels(),
	}
}

func (r *resourceMeta) GetName() string {
	return r.name
}

func (r *resourceMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (r *resourceMeta) GetVersion() string {
	return ""
}

func (r *resourceMeta) GetMesh() string {
	return r.mesh
}

func (r *resourceMeta) GetCreationTime() time.Time {
	return time.Unix(0, 0)
}

func (r *resourceMeta) GetModificationTime() time.Time {
	return time.Unix(0, 0)
}

func (r *resourceMeta) GetLabels() map[string]string {
	return maps.Clone(r.labels)
}
