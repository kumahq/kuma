package meshroute_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

func meshCtxWith(ms *meshservice_api.MeshServiceResource) xds_context.MeshContext {
	return xds_context.MeshContext{
		BaseMeshContext: &xds_context.BaseMeshContext{
			DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}),
		},
	}
}

var _ = Describe("SniForBackendRef", func() {
	DescribeTable("should return SNI and true when port is found",
		func(sectionName string) {
			ms := builders.MeshService().
				WithName("backend").
				WithMesh("default").
				AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
				Build()

			id := kri.WithSectionName(kri.From(ms), sectionName)
			ref := &resolve.RealResourceBackendRef{Resource: id}

			sni, ok := meshroute.SniForBackendRef(ref, meshCtxWith(ms), "")

			Expect(ok).To(BeTrue())
			Expect(sni).NotTo(BeEmpty())
			Expect(sni).To(ContainSubstring(".8080."))
		},
		Entry("by port name", "http"),
		Entry("by port value", "8080"),
	)

	It("returns empty string and false when sectionName does not match any port", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
			Build()

		id := kri.WithSectionName(kri.From(ms), "grpc")
		ref := &resolve.RealResourceBackendRef{Resource: id}

		sni, ok := meshroute.SniForBackendRef(ref, meshCtxWith(ms), "")

		Expect(ok).To(BeFalse())
		Expect(sni).To(BeEmpty())
	})

	It("returns empty string and false when service is not in the index", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
			Build()

		// empty index — service not registered
		emptyCtx := xds_context.MeshContext{
			BaseMeshContext: &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex(),
			},
		}

		id := kri.WithSectionName(kri.From(ms), "http")
		ref := &resolve.RealResourceBackendRef{Resource: id}

		sni, ok := meshroute.SniForBackendRef(ref, emptyCtx, "")

		Expect(ok).To(BeFalse())
		Expect(sni).To(BeEmpty())
	})
})
