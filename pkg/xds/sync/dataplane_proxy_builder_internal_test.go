package sync

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("DataplaneProxyBuilder resolveVIPOutbounds", func() {
	DescribeTable("transparent proxy enabled without transparentProxying section",
		func(allowAllOutbound bool, expectedOutbounds int) {
			ms := builders.MeshService().
				WithName("backend").
				AddIntPort(9000, 9000, metadata.ProtocolHTTP).
				Build()
			meshContext := xds_context.MeshContext{
				BaseMeshContext: &xds_context.BaseMeshContext{
					DestinationIndex: xds_context.NewDestinationIndex([]core_model.Resource{ms}).WithAllowAllOutbound(allowAllOutbound),
				},
				VIPOutbounds: xds_types.Outbounds{{
					Address:  "240.0.0.1",
					Port:     9000,
					Resource: kri.WithSectionName(kri.From(ms), "9000"),
				}},
			}
			dp := builders.Dataplane().WithAddress("127.0.0.1").Build()
			Expect(dp.Spec.GetNetworking().GetTransparentProxying()).To(BeNil())

			outbounds := (&DataplaneProxyBuilder{}).resolveVIPOutbounds(meshContext, dp, true, false)

			Expect(outbounds).To(HaveLen(expectedOutbounds))
		},
		Entry("deny by default", false, 0),
		Entry("allow all when allowAllOutbound is set", true, 1),
	)
})
