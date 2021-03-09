package system

import (
	"errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ZoneInsightType model.ResourceType = "ZoneInsight"
)

var _ model.Resource = &ZoneInsightResource{}

type ZoneInsightResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.ZoneInsight
}

func NewZoneInsightResource() *ZoneInsightResource {
	return &ZoneInsightResource{
		Spec: &system_proto.ZoneInsight{},
	}
}

func (t *ZoneInsightResource) GetType() model.ResourceType {
	return ZoneInsightType
}

func (t *ZoneInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneInsightResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneInsightResource) SetSpec(spec model.ResourceSpec) error {
	value, ok := spec.(*system_proto.ZoneInsight)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = value
		return nil
	}
}

func (t *ZoneInsightResource) Validate() error {
	return nil
}

func (t *ZoneInsightResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &ZoneInsightResourceList{}

type ZoneInsightResourceList struct {
	Items      []*ZoneInsightResource
	Pagination model.Pagination
}

func (l *ZoneInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneInsightResourceList) GetItemType() model.ResourceType {
	return ZoneInsightType
}

func (l *ZoneInsightResourceList) NewItem() model.Resource {
	return NewZoneInsightResource()
}

func (l *ZoneInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneInsightResource)(nil), r)
	}
}

func (l *ZoneInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewZoneInsightResource())
	registry.RegistryListType(&ZoneInsightResourceList{})
}
