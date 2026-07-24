package meshroute_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
)

var _ = Describe("SniForBackendRef", func() {
	DescribeTable("returns SNI built from resolved port",
		func(sectionName string) {
			ms := builders.MeshService().
				WithName("backend").
				WithMesh("default").
				AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
				Build()

			port, ok := ms.FindPortByName(sectionName)
			Expect(ok).To(BeTrue())

			id := kri.WithSectionName(kri.From(ms), sectionName)
			ref := &resolve.RealResourceBackendRef{Resource: id}

			sni := meshroute.SniForBackendRef(ref, ms, port, "")

			Expect(sni).NotTo(BeEmpty())
			Expect(sni).To(ContainSubstring(".8080."))
		},
		Entry("by port name", "http"),
		Entry("by port value", "8080"),
	)

	It("uses SNIName for MeshService destination", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
			Build()

		port, ok := ms.FindPortByName("http")
		Expect(ok).To(BeTrue())

		id := kri.WithSectionName(kri.From(ms), "http")
		id.ResourceType = meshservice_api.MeshServiceType
		ref := &resolve.RealResourceBackendRef{Resource: id}

		sni := meshroute.SniForBackendRef(ref, ms, port, "kuma-system")

		Expect(sni).To(ContainSubstring(ms.SNIName("kuma-system")))
	})
})
