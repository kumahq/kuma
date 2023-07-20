package etcd

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"time"
)

type resourceObject struct {
	MetaData []byte `json:"meta_data"`
	SpecData []byte `json:"spec_data"`
	Owner    struct {
		Type core_model.ResourceType `json:"type"`
		Mesh string                  `json:"mesh"`
		Name string                  `json:"name"`
	} `json:"owner"`
}

type etcdResourceMetaObject struct {
	Name             string    `json:"name"`
	Version          string    `json:"version"`
	Mesh             string    `json:"mesh"`
	CreationTime     time.Time `json:"creation_time"`
	ModificationTime time.Time `json:"modification_time"`
}

var _ core_model.ResourceMeta = &etcdResourceMetaObject{}

func (r *etcdResourceMetaObject) GetName() string {
	return r.Name
}

func (r *etcdResourceMetaObject) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}

func (r *etcdResourceMetaObject) GetVersion() string {
	return r.Version
}

func (r *etcdResourceMetaObject) GetMesh() string {
	return r.Mesh
}

func (r *etcdResourceMetaObject) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r *etcdResourceMetaObject) GetModificationTime() time.Time {
	return r.ModificationTime
}

func ToEtcdResourceMetaObject(resourceMeta core_model.ResourceMeta) core_model.ResourceMeta {
	if resourceMeta == nil {
		return &etcdResourceMetaObject{}
	}
	return &etcdResourceMetaObject{
		Name:             resourceMeta.GetName(),
		Version:          resourceMeta.GetVersion(),
		Mesh:             resourceMeta.GetMesh(),
		CreationTime:     resourceMeta.GetCreationTime(),
		ModificationTime: resourceMeta.GetModificationTime(),
	}
}
