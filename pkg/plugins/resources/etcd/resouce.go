package etcd

import (
	"time"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type resourceObject struct {
	MetaData []byte `json:"meta_data"`
	SpecData []byte `json:"spec_data"`
}
type indexResourceObject struct {
	EtcdResourceMetaObject etcdResourceMetaObject  `json:"etcd_resource_meta_object"`
	Type                   core_model.ResourceType `json:"type"`
}

type Owner struct {
	Type core_model.ResourceType `json:"type"`
	Mesh string                  `json:"mesh"`
	Name string                  `json:"name"`
}

type GetOwner interface {
	GetOwner() *Owner
}

type etcdResourceMetaObject struct {
	Name             string    `json:"name"`
	Version          string    `json:"version"`
	Mesh             string    `json:"mesh"`
	CreationTime     time.Time `json:"creation_time"`
	ModificationTime time.Time `json:"modification_time"`
	Owner            *Owner    `json:"owner"`
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

func (r *etcdResourceMetaObject) GetOwner() *Owner {
	return r.Owner
}

func ToEtcdResourceMetaObject(resourceMeta core_model.ResourceMeta, owner *Owner) core_model.ResourceMeta {
	if resourceMeta == nil {
		return &etcdResourceMetaObject{}
	}
	return &etcdResourceMetaObject{
		Name:             resourceMeta.GetName(),
		Version:          resourceMeta.GetVersion(),
		Mesh:             resourceMeta.GetMesh(),
		CreationTime:     resourceMeta.GetCreationTime(),
		ModificationTime: resourceMeta.GetModificationTime(),
		Owner:            owner,
	}
}

func NewIndexResource() *IndexResource {
	return &IndexResource{}
}

type IndexResource struct {
	typ                    core_model.ResourceType
	indexKey               string
	etcdResourceMetaObject core_model.ResourceMeta
}

func (i *IndexResource) GetMeta() core_model.ResourceMeta {
	return i.etcdResourceMetaObject
}

func (i *IndexResource) SetMeta(meta core_model.ResourceMeta) {
	i.etcdResourceMetaObject = meta
}

func (i *IndexResource) GetSpec() core_model.ResourceSpec {
	return nil
}

func (i *IndexResource) SetSpec(spec core_model.ResourceSpec) error {
	return nil
}

func (i *IndexResource) SetType(typ core_model.ResourceType) {
	i.typ = typ
}

func (i *IndexResource) SetIndexKey(indexKey string) {
	i.indexKey = indexKey
}

func (i *IndexResource) Descriptor() core_model.ResourceTypeDescriptor {
	return core_model.ResourceTypeDescriptor{
		Name: i.typ,
	}
}

func NewIndexResourceList() core_model.ResourceList {
	return &IndexResourceList{}
}

type IndexResourceList struct {
	items []core_model.Resource
}

func (i *IndexResourceList) GetItemType() core_model.ResourceType {
	return "IndexResourceList"
}

func (i *IndexResourceList) GetItems() []core_model.Resource {
	return i.items
}

func (i *IndexResourceList) NewItem() core_model.Resource {
	return NewIndexResource()
}

func (i *IndexResourceList) AddItem(resource core_model.Resource) error {
	if trr, ok := resource.(*IndexResource); ok {
		i.items = append(i.items, trr)
		return nil
	} else {
		return core_model.ErrorInvalidItemType((*IndexResource)(nil), resource)
	}
}

func (i *IndexResourceList) GetPagination() *core_model.Pagination {
	return &core_model.Pagination{}
}
