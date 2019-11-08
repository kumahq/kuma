package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/Kong/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
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
				inputType:    &proto.Mesh{},
				expectedType: &Mesh{},
				expectedKind: "Mesh",
			}),
			Entry("Dataplane", testCase{
				inputType:    &proto.Dataplane{},
				expectedType: &Dataplane{},
				expectedKind: "Dataplane",
			}),
			Entry("DataplaneInsight", testCase{
				inputType:    &proto.DataplaneInsight{},
				expectedType: &DataplaneInsight{},
				expectedKind: "DataplaneInsight",
			}),
			Entry("ProxyTemplate", testCase{
				inputType:    &proto.ProxyTemplate{},
				expectedType: &ProxyTemplate{},
				expectedKind: "ProxyTemplate",
			}),
			Entry("TrafficPermission", testCase{
				inputType:    &proto.TrafficPermission{},
				expectedType: &TrafficPermission{},
				expectedKind: "TrafficPermission",
			}),
			Entry("TrafficLog", testCase{
				inputType:    &proto.TrafficLog{},
				expectedType: &TrafficLog{},
				expectedKind: "TrafficLog",
			}),
			Entry("TrafficRoute", testCase{
				inputType:    &proto.TrafficRoute{},
				expectedType: &TrafficRoute{},
				expectedKind: "TrafficRoute",
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
				inputType:    &proto.Mesh{},
				expectedType: &MeshList{},
				expectedKind: "MeshList",
			}),
			Entry("DataplaneList", testCase{
				inputType:    &proto.Dataplane{},
				expectedType: &DataplaneList{},
				expectedKind: "DataplaneList",
			}),
			Entry("DataplaneInsightList", testCase{
				inputType:    &proto.DataplaneInsight{},
				expectedType: &DataplaneInsightList{},
				expectedKind: "DataplaneInsightList",
			}),
			Entry("ProxyTemplateList", testCase{
				inputType:    &proto.ProxyTemplate{},
				expectedType: &ProxyTemplateList{},
				expectedKind: "ProxyTemplateList",
			}),
			Entry("TrafficPermissionList", testCase{
				inputType:    &proto.TrafficPermission{},
				expectedType: &TrafficPermissionList{},
				expectedKind: "TrafficPermissionList",
			}),
			Entry("TrafficLogList", testCase{
				inputType:    &proto.TrafficLog{},
				expectedType: &TrafficLogList{},
				expectedKind: "TrafficLogList",
			}),
			Entry("TrafficRouteList", testCase{
				inputType:    &proto.TrafficRoute{},
				expectedType: &TrafficRouteList{},
				expectedKind: "TrafficRouteList",
			}),
		)
	})
})
