package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("ResolveTargetRef", func() {
	var resources context.Resources
	BeforeEach(func() {
		resources = context.NewResources()
	})

	addResource := func(resource core_model.Resource) {
		if _, ok := resources.MeshLocalResources[resource.Descriptor().Name]; !ok {
			list, err := registry.Global().NewList(resource.Descriptor().Name)
			Expect(err).ToNot(HaveOccurred())
			resources.MeshLocalResources[resource.Descriptor().Name] = list
		}
		Expect(resources.MeshLocalResources[resource.Descriptor().Name].AddItem(resource)).To(Succeed())
	}

	It("should resolve MeshService targetRef", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:      common_api.MeshService,
			Name:      pointer.To("backend"),
			Namespace: pointer.To("kuma-demo"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend"))
	})

	It("should not resolve MeshService targetRef when there is no MeshService", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:      common_api.MeshService,
			Name:      pointer.To("backend"),
			Namespace: pointer.To("kuma-demo"),
		}

		// given no MeshServices

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(BeEmpty())
	})

	It("should resolve MeshService targetRef with section name", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Name:        pointer.To("backend"),
			Namespace:   pointer.To("kuma-demo"),
			SectionName: pointer.To("tcp-port"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/tcp-port"))
	})

	It("should not resolve MeshService targetRef with section name that doesn't exist", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Name:        pointer.To("backend"),
			Namespace:   pointer.To("kuma-demo"),
			SectionName: pointer.To("tcp-port"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPort(8080, 8081, mesh.ProtocolTCP).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(BeEmpty())
	})

	It("should not resolve MeshService targetRef with section name being a port value and MeshService's port name is set", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Name:        pointer.To("backend"),
			Namespace:   pointer.To("kuma-demo"),
			SectionName: pointer.To("8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(BeEmpty())
	})

	It("should resolve MeshService targetRef with section name being a port value and MeshService's port name is unset", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshService,
			Name:        pointer.To("backend"),
			Namespace:   pointer.To("kuma-demo"),
			SectionName: pointer.To("8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPort(8080, 8081, mesh.ProtocolTCP).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/8080"))
	})

	It("should resolve MeshService targetRef with labels", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			},
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend"))
	})

	It("should resolve MeshService targetRef with labels and section name", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			},
			SectionName: pointer.To("tcp-port"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/tcp-port"))
	})

	It("should resolve MeshService targetRef with labels and section name being a port value and port's name unset", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			},
			SectionName: pointer.To("8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPort(8080, 8081, mesh.ProtocolTCP).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/8080"))
	})

	It("should not resolve MeshService targetRef with labels and section name being a port value and port's name set", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			},
			SectionName: pointer.To("8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(BeEmpty())
	})

	It("should not resolve MeshService targetRef with labels and section name that doesn't exist", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Labels: &map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			},
			SectionName: pointer.To("non-existent-section"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(BeEmpty())
	})

	It("should resolve legacy MeshService targetRef", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend_kuma-demo_svc_8080'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Name: pointer.To("backend_kuma-demo_svc_8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPortWithName(8080, 8081, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/tcp-port"))
	})

	It("should resolve legacy MeshService targetRef, MeshService's port without name", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend_kuma-demo_svc_8080'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Name: pointer.To("backend_kuma-demo_svc_8080"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPort(8080, 8081, mesh.ProtocolTCP).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend:section/8080"))
	})

	It("should resolve legacy MeshService targetRef for service less", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend_kuma-demo_svc'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind: common_api.MeshService,
			Name: pointer.To("backend_kuma-demo_svc"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshService().
			WithName("backend").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend",
			}).
			AddIntPort(8080, 8081, mesh.ProtocolTCP).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshservice:mesh/mesh-1:namespace/kuma-demo:name/backend"))
	})

	It("should resolve MeshMultiZoneService targetRef with section name", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'backend-mzsvc'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:        common_api.MeshMultiZoneService,
			Name:        pointer.To("backend-mzsvc"),
			Namespace:   pointer.To("kuma-demo"),
			SectionName: pointer.To("tcp-port"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshMultiZoneService().
			WithName("backend-mzsvc").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "backend-mzsvc",
			}).
			WithServiceLabelSelector(map[string]string{
				mesh_proto.DisplayName: "backend",
			}).
			AddIntPortWithName(8080, mesh.ProtocolTCP, "tcp-port").
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshmultizoneservice:mesh/mesh-1:namespace/kuma-demo:name/backend-mzsvc:section/tcp-port"))
	})

	It("should resolve MeshExternalService targetRef", func() {
		// given 'policy-1' with targetRef (somewhere in its spec) to 'mes'
		policyMeta := &test_model.ResourceMeta{
			Name: "policy-1",
			Mesh: "mesh-1",
		}
		targetRef := common_api.TargetRef{
			Kind:      common_api.MeshExternalService,
			Name:      pointer.To("mes"),
			Namespace: pointer.To("kuma-demo"),
		}
		// given actual MeshService 'backend' in 'kuma-demo' namespace with port 8080
		addResource(builders.MeshExternalService().
			WithName("backend-mzsvc").
			WithMesh("mesh-1").
			WithLabels(map[string]string{
				"k8s.kuma.io/namespace": "kuma-demo",
				"kuma.io/display-name":  "mes",
			}).
			Build(),
		)

		// when
		resolved := common.ResolveTargetRef(targetRef, policyMeta, resources)

		// then
		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Identifier().String()).To(Equal("meshexternalservice:mesh/mesh-1:namespace/kuma-demo:name/mes"))
	})
})
