package system

import (
	"errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	ZoneOverviewType model.ResourceType = "ZoneOverview"
)

var _ model.Resource = &ZoneOverviewResource{}

type ZoneOverviewResource struct {
	Meta model.ResourceMeta
	Spec *system_proto.ZoneOverview
}

func NewZoneOverviewResource() *ZoneOverviewResource {
	return &ZoneOverviewResource{
		Spec: &system_proto.ZoneOverview{},
	}
}

func (t *ZoneOverviewResource) GetType() model.ResourceType {
	return ZoneOverviewType
}

func (t *ZoneOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneOverviewResource) SetSpec(spec model.ResourceSpec) error {
	zoneOverview, ok := spec.(*system_proto.ZoneOverview)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = zoneOverview
		return nil
	}
}

func (t *ZoneOverviewResource) Validate() error {
	return nil
}

func (t *ZoneOverviewResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &ZoneOverviewResourceList{}

type ZoneOverviewResourceList struct {
	Items      []*ZoneOverviewResource
	Pagination model.Pagination
}

func (l *ZoneOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneOverviewResourceList) GetItemType() model.ResourceType {
	return ZoneOverviewType
}

func (l *ZoneOverviewResourceList) NewItem() model.Resource {
	return NewZoneOverviewResource()
}

func (l *ZoneOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneOverviewResource)(nil), r)
	}
}

func (l *ZoneOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func NewZoneOverviews(zones ZoneResourceList, insights ZoneInsightResourceList) ZoneOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*ZoneInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*ZoneOverviewResource
	for _, zone := range zones.Items {
		overview := ZoneOverviewResource{
			Meta: zone.Meta,
			Spec: &system_proto.ZoneOverview{
				Zone:        zone.Spec,
				ZoneInsight: nil,
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.ZoneInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return ZoneOverviewResourceList{
		Pagination: zones.Pagination,
		Items:      items,
	}
}
