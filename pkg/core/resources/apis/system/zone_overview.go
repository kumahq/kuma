package system

import (
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

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
