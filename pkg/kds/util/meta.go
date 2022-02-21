package util

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// KDS ResourceMeta only contains name and mesh.
// The rest is managed by the receiver of resources anyways. See ResourceSyncer#Sync
type resourceMeta struct {
	name string
	mesh string
}

func NewResourceMeta(name, mesh, version string, creationTime, modificationTime time.Time) model.ResourceMeta {
	return &resourceMeta{
		name: name,
		mesh: mesh,
	}
}

func CloneResourceMetaWithNewName(meta model.ResourceMeta, name string) model.ResourceMeta {
	return &resourceMeta{
		name: name,
		mesh: meta.GetMesh(),
	}
}

func kumaResourceMetaToResourceMeta(meta *mesh_proto.KumaResource_Meta) model.ResourceMeta {
	return &resourceMeta{
		name: meta.Name,
		mesh: meta.Mesh,
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
