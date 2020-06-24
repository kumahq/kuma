package util

import (
	"time"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/util/proto"
)

type resourceMeta struct {
	name             string
	version          string
	mesh             string
	creationTime     *time.Time
	modificationTime *time.Time
}

func ResourceKeyToMeta(name, mesh string) model.ResourceMeta {
	return &resourceMeta{
		name: name,
		mesh: mesh,
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
