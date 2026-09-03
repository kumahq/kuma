package resolve_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	"github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("Resolve TargetRef", func() {
	var resources context.Resources
	BeforeEach(func() {
		resources = context.NewResources()
	})

	targetLabels := func(name string) *map[string]string {
		return &map[string]string{
			mesh_proto.KubeNamespaceTag: "kuma-demo",
			mesh_proto.DisplayName:      name,
		}
	}

	addResource := func(resource core_model.Resource) {
		if _, ok := resources.MeshLocalResources[resource.Descriptor().Name]; !ok {
			list, err := registry.Global().NewList(resource.Descriptor().Name)
			Expect(err).ToNot(HaveOccurred())
			resources.MeshLocalResources[resource.Descriptor().Name] = list
		}
		Expect(resources.MeshLocalResources[resource.Descriptor().Name].AddItem(resource)).To(Succeed())
	}

	It("should resolve MeshService targetRef by labels", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:   common_api.MeshService,
			Labels: targetLabels("backend"),
		}
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend",
			}).
			AddIntPortWithName(8080, 8081, core_meta.ProtocolTCP, "tcp-port").
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("kri_msvc_mesh-1__kuma-demo_backend_"))
	})

	It("should not resolve MeshService targetRef when there is no MeshService", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:   common_api.MeshService,
			Labels: targetLabels("backend"),
		}

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(BeEmpty())
	})

	It("should resolve MeshService targetRef with section name", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Labels:      targetLabels("backend"),
			SectionName: pointer.To("tcp-port"),
		}
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend",
			}).
			AddIntPortWithName(8080, 8081, core_meta.ProtocolTCP, "tcp-port").
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("kri_msvc_mesh-1__kuma-demo_backend_tcp-port"))
	})

	It("should not resolve MeshService targetRef with section name that doesn't exist", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Labels:      targetLabels("backend"),
			SectionName: pointer.To("tcp-port"),
		}
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend",
			}).
			AddIntPort(8080, 8081, core_meta.ProtocolTCP).
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(BeEmpty())
	})

	It("should not resolve MeshService targetRef with section name being a port value and MeshService's port name is set", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Labels:      targetLabels("backend"),
			SectionName: pointer.To("8080"),
		}
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend",
			}).
			AddIntPortWithName(8080, 8081, core_meta.ProtocolTCP, "tcp-port").
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(BeEmpty())
	})

	It("should resolve MeshService targetRef with section name being a port value and MeshService's port name is unset", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Labels:      targetLabels("backend"),
			SectionName: pointer.To("8080"),
		}
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend",
			}).
			AddIntPort(8080, 8081, core_meta.ProtocolTCP).
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("kri_msvc_mesh-1__kuma-demo_backend_8080"))
	})

	It("should resolve MeshMultiZoneService targetRef with section name", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshMultiZoneService,
			Labels:      targetLabels("backend-mzsvc"),
			SectionName: pointer.To("tcp-port"),
		}
		addResource(builders.MeshMultiZoneService().
			WithName("backend-mzsvc").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "backend-mzsvc",
			}).
			WithServiceLabelSelector(map[string]string{
				mesh_proto.DisplayName: "backend",
			}).
			AddIntPortWithName(8080, core_meta.ProtocolTCP, "tcp-port").
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("kri_mzsvc_mesh-1__kuma-demo_backend-mzsvc_tcp-port"))
	})

	DescribeTable("should not resolve a stored targetRef of a removed kind", func(kind string) {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:   common_api.TargetRefKind(kind),
			Labels: targetLabels("backend"),
		}

		var resolved []*resolve.ResourceSection
		Expect(func() {
			resolved = resolve.TargetRef(targetRef, policyMeta, resources)
		}).ToNot(Panic())
		Expect(resolved).To(BeEmpty())
	},
		Entry("MeshSubset", "MeshSubset"),
		Entry("MeshServiceSubset", "MeshServiceSubset"),
	)

	It("should resolve MeshExternalService targetRef by labels", func() {
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:   common_api.MeshExternalService,
			Labels: targetLabels("mes"),
		}
		addResource(builders.MeshExternalService().
			WithName("backend-mzsvc").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				mesh_proto.KubeNamespaceTag: "kuma-demo",
				mesh_proto.DisplayName:      "mes",
			}).
			Build(),
		)

		resolved := resolve.TargetRef(targetRef, policyMeta, resources)

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("kri_extsvc_mesh-1__kuma-demo_mes_"))
	})
})
