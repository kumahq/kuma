package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
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
				SpiffeId:         dataplaneSpiffeID(dataplane.Meta.GetMesh(), dataplane.Spec),
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

// dataplaneSpiffeID returns the SPIFFE ID for a dataplane based on its first inbound service
// or gateway service tag. Returns empty string when no service can be determined.
func dataplaneSpiffeID(mesh string, dp *mesh_proto.Dataplane) string {
	if dp == nil {
		return ""
	}
	var svc string
	if inbounds := dp.GetNetworking().GetInbound(); len(inbounds) > 0 {
		svc = inbounds[0].Tags[mesh_proto.ServiceTag]
	} else if gw := dp.GetNetworking().GetGateway(); gw != nil {
		svc = gw.Tags[mesh_proto.ServiceTag]
	}
	if svc == "" {
		return ""
	}
	return fmt.Sprintf("spiffe://%s/%s", mesh, svc)
}
