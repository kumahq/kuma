package util

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type resourceMeta struct {
	name             string
	version          string
	mesh             string
	creationTime     *time.Time
	modificationTime *time.Time
}

func NewResourceMeta(name, mesh, version string, creationTime, modificationTime time.Time) model.ResourceMeta {
	return &resourceMeta{
		name:             name,
		mesh:             mesh,
		version:          version,
		creationTime:     &creationTime,
		modificationTime: &modificationTime,
	}
}

func CloneResourceMetaWithNewName(meta model.ResourceMeta, name string) model.ResourceMeta {
	creationTime := meta.GetCreationTime()
	modificationTime := meta.GetModificationTime()

	return &resourceMeta{
		name:             name,
		version:          meta.GetVersion(),
		mesh:             meta.GetMesh(),
		creationTime:     &creationTime,
		modificationTime: &modificationTime,
	}
}

func kumaResourceMetaToResourceMeta(meta *mesh_proto.KumaResource_Meta) model.ResourceMeta {
	return &resourceMeta{
		name:             meta.Name,
		mesh:             meta.Mesh,
		version:          meta.Version,
		creationTime:     proto.MustTimestampFromProto(meta.CreationTime),
		modificationTime: proto.MustTimestampFromProto(meta.ModificationTime),
	}
}

func (r *resourceMeta) GetName() string {
	return r.name
}

func (r *resourceMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (r *resourceMeta) GetVersion() string {
	return r.version
}

func (r *resourceMeta) GetMesh() string {
	return r.mesh
}

func (r *resourceMeta) GetCreationTime() time.Time {
	return *r.creationTime
}

func (r *resourceMeta) GetModificationTime() time.Time {
	return *r.modificationTime
}
