package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/k8s/native/pkg/registry"
)

var _ = Describe("global TypeRegistry", func() {
	Context("object types", func() {
		type testCase struct {
			inputType    registry.ResourceType
			expectedType model.KubernetesObject
			expectedKind string
		}

		DescribeTable("should include all mesh types",
			func(given testCase) {
				// given
				expectedAPIVersion := GroupVersion

				// when
				obj, err := registry.Global().NewObject(given.inputType)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(obj).ToNot(BeNil())
				Expect(obj).To(BeAssignableToTypeOf(given.expectedType))
				// and
				Expect(obj.GetObjectKind().GroupVersionKind()).To(Equal(expectedAPIVersion.WithKind(given.expectedKind)))
			},
			Entry("Mesh", testCase{
				inputType:    &mesh_proto.Mesh{},
				expectedType: &Mesh{},
				expectedKind: "Mesh",
			}),
			Entry("Dataplane", testCase{
				inputType:    &mesh_proto.Dataplane{},
				expectedType: &Dataplane{},
				expectedKind: "Dataplane",
			}),
			Entry("DataplaneInsight", testCase{
				inputType:    &mesh_proto.DataplaneInsight{},
				expectedType: &DataplaneInsight{},
				expectedKind: "DataplaneInsight",
			}),
			Entry("ExternalService", testCase{
				inputType:    &mesh_proto.ExternalService{},
				expectedType: &ExternalService{},
				expectedKind: "ExternalService",
			}),
			Entry("ProxyTemplate", testCase{
				inputType:    &mesh_proto.ProxyTemplate{},
				expectedType: &ProxyTemplate{},
				expectedKind: "ProxyTemplate",
			}),
			Entry("RateLimit", testCase{
				inputType:    &mesh_proto.RateLimit{},
				expectedType: &RateLimit{},
				expectedKind: "RateLimit",
			}),
		)
	})

	Context("list types", func() {
		type testCase struct {
			inputType    registry.ResourceType
			expectedType model.KubernetesList
			expectedKind string
		}

		DescribeTable("should include all mesh types",
			func(given testCase) {
				// when
				obj, err := registry.Global().NewList(given.inputType)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(obj).ToNot(BeNil())
				Expect(obj).To(BeAssignableToTypeOf(given.expectedType))
				// and
				Expect(obj.GetObjectKind().GroupVersionKind()).To(Equal(GroupVersion.WithKind(given.expectedKind)))
			},
			Entry("MeshList", testCase{
				inputType:    &mesh_proto.Mesh{},
				expectedType: &MeshList{},
				expectedKind: "MeshList",
			}),
			Entry("DataplaneList", testCase{
				inputType:    &mesh_proto.Dataplane{},
				expectedType: &DataplaneList{},
				expectedKind: "DataplaneList",
			}),
			Entry("DataplaneInsightList", testCase{
				inputType:    &mesh_proto.DataplaneInsight{},
				expectedType: &DataplaneInsightList{},
				expectedKind: "DataplaneInsightList",
			}),
			Entry("ExternalServiceList", testCase{
				inputType:    &mesh_proto.ExternalService{},
				expectedType: &ExternalServiceList{},
				expectedKind: "ExternalServiceList",
			}),
			Entry("ProxyTemplateList", testCase{
				inputType:    &mesh_proto.ProxyTemplate{},
				expectedType: &ProxyTemplateList{},
				expectedKind: "ProxyTemplateList",
			}),
		)
	})
})
