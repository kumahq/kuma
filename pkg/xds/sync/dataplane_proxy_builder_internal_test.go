package sync

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

var _ = Describe("asOutbounds", func() {
	DescribeTable("names an explicit outbound section by port name",
		func(port uint32, expectedSection string) {
			ms := builders.MeshService().
				WithName("backend").
				AddIntPortWithName(80, 8080, core_meta.ProtocolHTTP, "http").
				AddIntPort(81, 8081, core_meta.ProtocolTCP).
				Build()
			meshContext := xds_context.MeshContext{
				BaseMeshContext: &xds_context.BaseMeshContext{
					DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}),
				},
			}
			dataplane := builders.Dataplane().
				AddOutbound(builders.Outbound().WithAddress("127.0.0.1").WithPort(10080).WithMeshService("backend", port)).
				Build()

			outbounds := asOutbounds(dataplane, meshContext.ResolveResourceIdentifier, meshContext.GetServiceByKRI)

			Expect(outbounds).To(HaveLen(1))
			Expect(outbounds[0].Resource).To(Equal(kri.WithSectionName(kri.From(ms), expectedSection)))
		},
		Entry("named port", uint32(80), "http"),
		Entry("unnamed port", uint32(81), "81"),
	)

	DescribeTable("names an explicit MeshExternalService outbound section by match port",
		func(port uint32) {
			mes := builders.MeshExternalService().WithName("ext").Build()
			meshContext := xds_context.MeshContext{
				BaseMeshContext: &xds_context.BaseMeshContext{
					DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{mes}),
				},
			}
			dataplane := builders.Dataplane().
				AddOutbound(builders.Outbound().WithAddress("127.0.0.1").WithPort(10080).WithMeshExternalService("ext", port)).
				Build()

			outbounds := asOutbounds(dataplane, meshContext.ResolveResourceIdentifier, meshContext.GetServiceByKRI)

			Expect(outbounds).To(HaveLen(1))
			Expect(outbounds[0].Resource).To(Equal(kri.WithSectionName(kri.From(mes), "9000")))
		},
		Entry("no port", uint32(0)),
		Entry("port other than match port", uint32(443)),
	)
})
