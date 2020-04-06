package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/test/resources/model"
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
				},
				{
					Meta: &model.ResourceMeta{
						Name: "dp2",
						Mesh: "mesh1",
					},
				},
			}}

			insights := DataplaneInsightResourceList{Items: []*DataplaneInsightResource{
				{
					Meta: &model.ResourceMeta{
						Name: "dp1",
						Mesh: "mesh1",
					},
				},
			}}

			overviews := NewDataplaneOverviews(dataplanes, insights)
			Expect(overviews.Items).To(HaveLen(2))
			Expect(overviews.Items[0].Spec.Dataplane).To(Equal(&dataplanes.Items[0].Spec))
			Expect(overviews.Items[0].Spec.DataplaneInsight).To(Equal(&insights.Items[0].Spec))
			Expect(overviews.Items[1].Spec.Dataplane).To(Equal(&dataplanes.Items[1].Spec))
			Expect(overviews.Items[1].Spec.DataplaneInsight).To(BeNil())
		})
	})

	Describe("RetainMatchingTags", func() {
		overviews := DataplaneOverviewResourceList{
			Items: []*DataplaneOverviewResource{
				{
					Spec: v1alpha1.DataplaneOverview{
						Dataplane: &v1alpha1.Dataplane{
							Networking: &v1alpha1.Dataplane_Networking{
								Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
									{
										Tags: map[string]string{
											"service": "mobile",
											"version": "v1",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		type testCase struct {
			tags     map[string]string
			expected DataplaneOverviewResourceList
		}
		DescribeTable("should retain overviews", func(given testCase) {
			// when
			overviews.RetainMatchingTags(given.tags)

			// then
			Expect(overviews).To(Equal(given.expected))
		},
			Entry("should retain all with empty map", testCase{
				tags:     map[string]string{},
				expected: overviews,
			}),
			Entry("should retain with one matching tag", testCase{
				tags:     map[string]string{"service": "mobile"},
				expected: overviews,
			}),
			Entry("should retain with matching all tags", testCase{
				tags:     map[string]string{"service": "mobile", "version": "v1"},
				expected: overviews,
			}),
			Entry("should retain none with mismatching tag", testCase{
				tags:     map[string]string{"service": "mobile", "version": "v2"},
				expected: DataplaneOverviewResourceList{Items: []*DataplaneOverviewResource{}},
			}))
	})
	Describe("RetainGateWayDataPlanes", func() {
		dataplanes := DataplaneOverviewResourceList{
			Items: []*DataplaneOverviewResource{
				{
					Spec: v1alpha1.DataplaneOverview{
						Dataplane: &v1alpha1.Dataplane{
							Networking: &v1alpha1.Dataplane_Networking{
								Gateway: &v1alpha1.Dataplane_Networking_Gateway{
									Tags: map[string]string{
										"service": "gateway",
									},
								},
							},
						},
					},
				},
				{
					Spec: v1alpha1.DataplaneOverview{
						Dataplane: &v1alpha1.Dataplane{
							Networking: &v1alpha1.Dataplane_Networking{
								Inbound: []*v1alpha1.Dataplane_Networking_Inbound{
									{
										Tags: map[string]string{
											"service": "mobile",
											"version": "v1",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		gatewayDataplanes := DataplaneOverviewResourceList{
			Items: []*DataplaneOverviewResource{
				{
					Spec: v1alpha1.DataplaneOverview{
						Dataplane: &v1alpha1.Dataplane{
							Networking: &v1alpha1.Dataplane_Networking{
								Gateway: &v1alpha1.Dataplane_Networking_Gateway{
									Tags: map[string]string{
										"service": "gateway",
									},
								},
							},
						},
					},
				},
			},
		}
		It("should retain gateway overviews", func() {
			dataplanes.RetainGatewayDataplanes()
			Expect(dataplanes).To(Equal(gatewayDataplanes))
		})
	})
})
