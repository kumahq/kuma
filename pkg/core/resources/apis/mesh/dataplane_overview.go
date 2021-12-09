package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func NewDataplaneOverviews(dataplanes DataplaneResourceList, insights DataplaneInsightResourceList) DataplaneOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*DataplaneInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*DataplaneOverviewResource
	for _, dataplane := range dataplanes.Items {
		overview := DataplaneOverviewResource{
			Meta: dataplane.Meta,
			Spec: &mesh_proto.DataplaneOverview{
				Dataplane:        dataplane.Spec,
				DataplaneInsight: nil,
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.DataplaneInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return DataplaneOverviewResourceList{
		Pagination: dataplanes.Pagination,
		Items:      items,
	}
}
