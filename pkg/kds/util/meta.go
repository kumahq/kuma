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

func CloneResourceMetaWithNewName(meta model.ResourceMeta, name string) model.ResourceMeta {
	labels := maps.Clone(meta.GetLabels())
	if labels == nil {
		labels = map[string]string{}
	}
	if labels[mesh_proto.DisplayName] == "" {
		labels[mesh_proto.DisplayName] = meta.GetName()
	}
	return &resourceMeta{
		name:   name,
		mesh:   meta.GetMesh(),
		labels: labels,
	}
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
	return r.labels
}
