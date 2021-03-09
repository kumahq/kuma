package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("DataplaneOverview", func() {

	Describe("NewDataplaneOverviews", func() {
		It("should create overviews from dataplanes and insights", func() {
			dataplanes := DataplaneResourceList{Items: []*DataplaneResource{
				{
					Meta: &model.ResourceMeta{
						Name: "dp1",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.Dataplane{},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "dp2",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.Dataplane{},
				},
			}}

			insights := DataplaneInsightResourceList{Items: []*DataplaneInsightResource{
				{
					Meta: &model.ResourceMeta{
						Name: "dp1",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.DataplaneInsight{},
				},
			}}

			overviews := NewDataplaneOverviews(dataplanes, insights)
			Expect(overviews.Items).To(HaveLen(2))
			Expect(overviews.Items[0].Spec.Dataplane).To(MatchProto(dataplanes.Items[0].Spec))
			Expect(overviews.Items[0].Spec.DataplaneInsight).To(MatchProto(insights.Items[0].Spec))
			Expect(overviews.Items[1].Spec.Dataplane).To(MatchProto(dataplanes.Items[1].Spec))
			Expect(overviews.Items[1].Spec.DataplaneInsight).To(BeNil())
		})
	})
})
