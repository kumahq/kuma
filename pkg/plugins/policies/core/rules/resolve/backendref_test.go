package resolve_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("Resolve BackendRef", func() {
	DescribeTable("should resolve label-selected real backendRefs",
		func(kind common_api.TargetRefKind, labels map[string]string, expectedType core_model.ResourceType) {
			origin := kri.Identifier{Mesh: "mesh-1", Namespace: "kuma-demo"}
			expected := kri.Identifier{ResourceType: expectedType, Mesh: "mesh-1", Name: "resolved"}

			var capturedType core_model.ResourceType
			var capturedLabels map[string]string
			resolved, ok := resolve.BackendRef(origin, common_api.BackendRef{
				TargetRef: common_api.TargetRef{
					Kind:   kind,
					Labels: pointer.To(labels),
				},
			}, func(resType core_model.ResourceType, gotLabels map[string]string) kri.Identifier {
				capturedType = resType
				capturedLabels = gotLabels
				return expected
			})

			Expect(ok).To(BeTrue())
			Expect(resolved.LegacyBackendRef()).To(BeNil())
			Expect(resolved.RealResourceBackendRef()).ToNot(BeNil())
			Expect(resolved.Resource()).To(Equal(expected))
			Expect(capturedType).To(Equal(expectedType))
			Expect(capturedLabels).To(Equal(labels))
		},
		Entry("MeshService", common_api.MeshService, map[string]string{
			mesh_proto.DisplayName:      "backend",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
		}, core_model.ResourceType(common_api.MeshService)),
		Entry("MeshExternalService", common_api.MeshExternalService, map[string]string{
			mesh_proto.DisplayName: "payments",
		}, core_model.ResourceType(common_api.MeshExternalService)),
		Entry("MeshMultiZoneService", common_api.MeshMultiZoneService, map[string]string{
			mesh_proto.DisplayName: "global-backend",
		}, core_model.ResourceType(common_api.MeshMultiZoneService)),
	)

	It("should treat MeshService backendRefs without port or sectionName as real resources when labels identify them", func() {
		origin := kri.Identifier{Mesh: "mesh-1", Namespace: "kuma-demo"}
		expected := kri.Identifier{ResourceType: core_model.ResourceType(common_api.MeshService), Mesh: "mesh-1", Name: "backend"}

		resolved, ok := resolve.BackendRef(origin, common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Labels: pointer.To(map[string]string{
					mesh_proto.DisplayName:      "backend",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
				}),
			},
		}, func(_ core_model.ResourceType, _ map[string]string) kri.Identifier {
			return expected
		})

		Expect(ok).To(BeTrue())
		Expect(resolved.ReferencesRealResource()).To(BeTrue())
		Expect(resolved.LegacyBackendRef()).To(BeNil())
		Expect(resolved.Resource()).To(Equal(expected))
	})

	It("should derive labels from a MeshService name and policy namespace before resolving", func() {
		origin := kri.Identifier{Mesh: "mesh-1", Namespace: "kuma-demo"}
		expected := kri.Identifier{ResourceType: core_model.ResourceType(common_api.MeshService), Mesh: "mesh-1", Name: "backend"}

		var capturedLabels map[string]string
		resolved, ok := resolve.BackendRef(origin, common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: pointer.To("backend"),
			},
		}, func(_ core_model.ResourceType, gotLabels map[string]string) kri.Identifier {
			capturedLabels = gotLabels
			return expected
		})

		Expect(ok).To(BeTrue())
		Expect(capturedLabels).To(Equal(map[string]string{
			mesh_proto.DisplayName:      "backend",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
		}))
		Expect(resolved.Resource()).To(Equal(expected))
	})

	It("should derive compatibility labels and port from legacy MeshService name forms", func() {
		origin := kri.Identifier{Mesh: "mesh-1", Namespace: "ignored"}
		expected := kri.Identifier{ResourceType: core_model.ResourceType(common_api.MeshService), Mesh: "mesh-1", Name: "backend"}

		var capturedLabels map[string]string
		resolved, ok := resolve.BackendRef(origin, common_api.BackendRef{
			TargetRef: common_api.TargetRef{
				Kind: common_api.MeshService,
				Name: pointer.To("backend_kuma-demo_svc_8080"),
			},
		}, func(_ core_model.ResourceType, gotLabels map[string]string) kri.Identifier {
			capturedLabels = gotLabels
			return expected
		})

		Expect(ok).To(BeTrue())
		Expect(capturedLabels).To(Equal(map[string]string{
			mesh_proto.DisplayName:      "backend",
			mesh_proto.KubeNamespaceTag: "kuma-demo",
		}))
		Expect(resolved.ReferencesRealResource()).To(BeTrue())
		Expect(resolved.Resource()).To(Equal(kri.WithSectionName(expected, "8080")))
	})
})
