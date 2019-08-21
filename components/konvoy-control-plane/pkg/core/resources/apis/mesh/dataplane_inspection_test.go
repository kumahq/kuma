package mesh

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("DataplaneInspection", func() {

	Describe("NewDataplaneInspections", func() {
		It("should create inspections from dataplanes and insights", func() {
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

			inspections := NewDataplaneInspections(dataplanes, insights)
			Expect(inspections.Items).To(HaveLen(2))
			Expect(inspections.Items[0].Spec.Dataplane).To(Equal(dataplanes.Items[0].Spec))
			Expect(inspections.Items[0].Spec.DataplaneInsight).To(Equal(insights.Items[0].Spec))
			Expect(inspections.Items[1].Spec.Dataplane).To(Equal(dataplanes.Items[1].Spec))
			Expect(inspections.Items[1].Spec.DataplaneInsight).To(Equal(v1alpha1.DataplaneInsight{}))
		})
	})

	Describe("RetainMatchingTags", func() {
		inspections := DataplaneInspectionResourceList{
			Items: []*DataplaneInspectionResource{
				{
					Spec: v1alpha1.DataplaneInspection{
						Dataplane: v1alpha1.Dataplane{
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
			expected DataplaneInspectionResourceList
		}
		DescribeTable("should retain inspections", func(given testCase) {
			// when
			inspections.RetainMatchingTags(given.tags)

			// then
			Expect(inspections).To(Equal(given.expected))
		},
			Entry("should retain all with empty map", testCase{
				tags:     map[string]string{},
				expected: inspections,
			}),
			Entry("should retain with one matching tag", testCase{
				tags:     map[string]string{"service": "mobile"},
				expected: inspections,
			}),
			Entry("should retain with matching all tags", testCase{
				tags:     map[string]string{"service": "mobile", "version": "v1"},
				expected: inspections,
			}),
			Entry("should retain none with mismatching tag", testCase{
				tags:     map[string]string{"service": "mobile", "version": "v2"},
				expected: DataplaneInspectionResourceList{Items: []*DataplaneInspectionResource{}},
			}))
	})
})
