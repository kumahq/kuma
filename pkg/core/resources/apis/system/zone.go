package system

import (
	"errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ZoneType model.ResourceType = "Zone"
)

var _ model.Resource = &ZoneResource{}

type ZoneResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.Zone
}

func NewZoneResource() *ZoneResource {
	return &ZoneResource{
		Spec: &system_proto.Zone{},
	}
}

func (t *ZoneResource) GetType() model.ResourceType {
	return ZoneType
}

func (t *ZoneResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*system_proto.Zone)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = value
		return nil
	}
}

func (t *ZoneResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &ZoneResourceList{}

type ZoneResourceList struct {
	Items      []*ZoneResource
	Pagination model.Pagination
}

func (l *ZoneResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneResourceList) GetItemType() model.ResourceType {
	return ZoneType
}

func (l *ZoneResourceList) NewItem() model.Resource {
	return NewZoneResource()
}

func (l *ZoneResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneResource)(nil), r)
	}
}

func (l *ZoneResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewZoneResource())
	registry.RegistryListType(&ZoneResourceList{})
}
