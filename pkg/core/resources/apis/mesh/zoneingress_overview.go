package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	ZoneIngressOverviewType model.ResourceType = "ZoneIngressOverview"
)

var _ model.Resource = &ZoneIngressOverviewResource{}

type ZoneIngressOverviewResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.ZoneIngressOverview
}

func NewZoneIngressOverviewResource() *ZoneIngressOverviewResource {
	return &ZoneIngressOverviewResource{
		Spec: &mesh_proto.ZoneIngressOverview{},
	}
}

func (t *ZoneIngressOverviewResource) GetType() model.ResourceType {
	return ZoneIngressOverviewType
}

func (t *ZoneIngressOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ZoneIngressOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ZoneIngressOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ZoneIngressOverviewResource) SetSpec(spec model.ResourceSpec) error {
	zoneIngressOverview, ok := spec.(*mesh_proto.ZoneIngressOverview)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = zoneIngressOverview
		return nil
	}
}

func (t *ZoneIngressOverviewResource) Validate() error {
	return nil
}

func (t *ZoneIngressOverviewResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &ZoneIngressOverviewResourceList{}

type ZoneIngressOverviewResourceList struct {
	Items      []*ZoneIngressOverviewResource
	Pagination model.Pagination
}

func (l *ZoneIngressOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ZoneIngressOverviewResourceList) GetItemType() model.ResourceType {
	return ZoneIngressOverviewType
}
func (l *ZoneIngressOverviewResourceList) NewItem() model.Resource {
	return NewZoneIngressOverviewResource()
}
func (l *ZoneIngressOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneIngressOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneIngressOverviewResource)(nil), r)
	}
}

func (l *ZoneIngressOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func NewZoneIngressOverviews(zoneIngresses ZoneIngressResourceList, insights ZoneIngressInsightResourceList) ZoneIngressOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*ZoneIngressInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*ZoneIngressOverviewResource
	for _, zoneIngress := range zoneIngresses.Items {
		overview := ZoneIngressOverviewResource{
			Meta: zoneIngress.Meta,
			Spec: &mesh_proto.ZoneIngressOverview{
				ZoneIngress:        zoneIngress.Spec,
				ZoneIngressInsight: nil,
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.ZoneIngressInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return ZoneIngressOverviewResourceList{
		Pagination: zoneIngresses.Pagination,
		Items:      items,
	}
}
