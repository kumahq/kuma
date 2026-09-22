package sync_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	"github.com/kumahq/kuma/v3/pkg/xds/sync"
)

var _ = Describe("ResolveOutbounds", func() {
	DescribeTable("names an explicit outbound section by port name",
		func(port uint32, expectedSection string) {
			ms := builders.MeshService().
				WithName("backend").
				AddIntPortWithName(80, 8080, core_meta.ProtocolHTTP, "http").
				AddIntPort(81, 8081, core_meta.ProtocolTCP).
				Build()
			base := &xds_context.BaseMeshContext{
				DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}),
			}
			dataplane := builders.Dataplane().
				AddOutbound(builders.Outbound().WithAddress("127.0.0.1").WithPort(10080).WithMeshService("backend", port)).
				Build()

			outbounds := sync.ResolveOutbounds(base, dataplane, false, false)

			Expect(outbounds).To(HaveLen(1))
			Expect(outbounds[0].Resource).To(Equal(kri.WithSectionName(kri.From(ms), expectedSection)))
		},
		Entry("named port", uint32(80), "http"),
		Entry("unnamed port", uint32(81), "81"),
	)
})
