// +kubebuilder:object:generate=true
package v1alpha1

import (
	"errors"
	"fmt"

	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// ZoneOverview defines the projected state of a Zone. It is composed per request from a
// Zone and its insight, never stored and never synced, so it is written by hand rather
// than generated: the resource generator produces one resource per package, and a
// generated overview would have to import the Zone package that imports it back.
//
// Zone is always populated, see newZoneOverview. A nil one would drop the field from the
// response without any error: omitempty skips it, and the inlined REST representation
// throws the whole spec away once it renders as an empty object.
type ZoneOverview struct {
	Zone        *Zone                        `json:"zone,omitempty"`
	ZoneInsight *zoneinsight_api.ZoneInsight `json:"zoneInsight,omitempty"`
}

func newZoneOverview(zone *Zone) *ZoneOverview {
	if zone == nil {
		zone = &Zone{}
	}
	return &ZoneOverview{Zone: zone}
}

const ZoneOverviewType model.ResourceType = "ZoneOverview"

var (
	_ model.Resource         = &ZoneOverviewResource{}
	_ model.OverviewResource = &ZoneOverviewResource{}
	_ model.ResourceList     = &ZoneOverviewResourceList{}
)

type ZoneOverviewResource struct {
	Meta model.ResourceMeta
	Spec *ZoneOverview
}

func NewZoneOverviewResource() *ZoneOverviewResource {
	return &ZoneOverviewResource{Spec: &ZoneOverview{}}
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
	specType, ok := spec.(*ZoneOverview)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	}
	if specType == nil {
		t.Spec = &ZoneOverview{}
	} else {
		t.Spec = specType
	}
	return nil
}

func (t *ZoneOverviewResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *ZoneOverviewResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *ZoneOverviewResource) Descriptor() model.ResourceTypeDescriptor {
	return ZoneOverviewResourceTypeDescriptor
}

func (t *ZoneOverviewResource) SetOverviewSpec(resource model.Resource, insight model.Resource) error {
	t.SetMeta(resource.GetMeta())
	zone, ok := resource.GetSpec().(*Zone)
	if !ok {
		return errors.New("failed to convert to resource type 'Zone'")
	}
	overview := newZoneOverview(zone)
	if insight != nil {
		ins, ok := insight.GetSpec().(*zoneinsight_api.ZoneInsight)
		if !ok {
			return errors.New("failed to convert to insight type 'ZoneInsight'")
		}
		overview.ZoneInsight = ins
	}
	return t.SetSpec(overview)
}

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
	}
	return model.ErrorInvalidItemType((*ZoneOverviewResource)(nil), r)
}

func (l *ZoneOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *ZoneOverviewResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var ZoneOverviewResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                  ZoneOverviewType,
	Resource:              NewZoneOverviewResource(),
	ResourceList:          &ZoneOverviewResourceList{},
	ReadOnly:              false,
	AdminOnly:             false,
	Scope:                 model.ScopeGlobal,
	WsPath:                "",
	KumactlArg:            "",
	KumactlListArg:        "",
	AllowToInspect:        false,
	IsPolicy:              false,
	SingularDisplayName:   "Zone Overview",
	PluralDisplayName:     "Zone Overviews",
	IsExperimental:        false,
	IsProxy:               false,
	AffectsPolicyMatching: false,
}

func NewZoneOverviews(zones ZoneResourceList, insights zoneinsight_api.ZoneInsightResourceList) ZoneOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*zoneinsight_api.ZoneInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*ZoneOverviewResource
	for _, zone := range zones.Items {
		overview := ZoneOverviewResource{
			Meta: zone.Meta,
			Spec: newZoneOverview(zone.Spec),
		}
		if insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]; exists {
			overview.Spec.ZoneInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return ZoneOverviewResourceList{
		Pagination: zones.Pagination,
		Items:      items,
	}
}
