package outbound_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
)

var _ = Describe("ResourceContext", func() {
	var mesh *core_mesh.MeshResource

	BeforeEach(func() {
		mesh = builders.Mesh().WithName("test-mesh").Build()
	})

	Describe("RootContext", func() {
		It("should create a ResourceContext with mesh identifier", func() {
			// when
			resourceRules := outbound.ResourceRules{
				kri.From(mesh, ""): outbound.ResourceRule{Conf: []interface{}{"mesh-conf"}},
			}
			rc := outbound.RootContext[string](mesh, resourceRules)

			// then
			Expect(rc).NotTo(BeNil())
			Expect(rc.Conf()).To(Equal("mesh-conf"))
		})
	})

	Describe("WithID", func() {
		It("should add a new identifier to the ResourceContext", func() {
			// given
			id := kri.Identifier{
				ResourceType: "TestResource",
				Mesh:         "test-mesh",
				Name:         "test-resource",
			}
			resourceRules := outbound.ResourceRules{
				kri.From(mesh, ""): outbound.ResourceRule{Conf: []interface{}{"mesh-conf"}},
				id:                 outbound.ResourceRule{Conf: []interface{}{"test-conf"}},
			}
			rc := outbound.RootContext[string](mesh, resourceRules)

			// when
			newRc := rc.WithID(id)

			// then
			Expect(newRc).NotTo(BeNil())
			Expect(newRc).NotTo(BeIdenticalTo(rc)) // Should be a new instance
			Expect(newRc.Conf()).To(Equal("test-conf"))
		})

		It("should return mesh's conf when there is no specific conf for the test-resource", func() {
			// given
			id := kri.Identifier{
				ResourceType: "TestResource",
				Mesh:         "test-mesh",
				Name:         "test-resource",
			}
			resourceRules := outbound.ResourceRules{
				kri.From(mesh, ""): outbound.ResourceRule{Conf: []interface{}{"mesh-conf"}},
			}
			rc := outbound.RootContext[string](mesh, resourceRules)

			// when
			newRc := rc.WithID(id)

			// then
			Expect(newRc).NotTo(BeNil())
			Expect(newRc).NotTo(BeIdenticalTo(rc)) // Should be a new instance
			Expect(newRc.Conf()).To(Equal("mesh-conf"))
		})
	})
})
