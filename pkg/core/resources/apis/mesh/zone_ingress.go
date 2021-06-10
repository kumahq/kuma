package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ZoneIngressType model.ResourceType = "ZoneIngress"
)

var _ model.Resource = &ZoneIngressResource{}

type ZoneIngressResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.ZoneIngress
}

func NewZoneIngressResource() *ZoneIngressResource {
	return &ZoneIngressResource{
		Spec: &mesh_proto.ZoneIngress{},
	}
}

func (r *ZoneIngressResource) GetType() model.ResourceType {
	return ZoneIngressType
}
func (r *ZoneIngressResource) GetMeta() model.ResourceMeta {
	return r.Meta
}
func (r *ZoneIngressResource) SetMeta(m model.ResourceMeta) {
	r.Meta = m
}
func (r *ZoneIngressResource) GetSpec() model.ResourceSpec {
	return r.Spec
}
func (r *ZoneIngressResource) SetSpec(spec model.ResourceSpec) error {
	zoneIngress, ok := spec.(*mesh_proto.ZoneIngress)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		r.Spec = zoneIngress
		return nil
	}
}

func (r *ZoneIngressResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &DataplaneResourceList{}

type ZoneIngressResourceList struct {
	Items      []*ZoneIngressResource
	Pagination model.Pagination
}

func (l *ZoneIngressResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneIngressResourceList) GetItemType() model.ResourceType {
	return ZoneIngressType
}
func (l *ZoneIngressResourceList) NewItem() model.Resource {
	return NewZoneIngressResource()
}
func (l *ZoneIngressResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneIngressResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneIngressResource)(nil), r)
	}
}
func (l *ZoneIngressResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewZoneIngressResource())
	registry.RegistryListType(&ZoneIngressResourceList{})
}
