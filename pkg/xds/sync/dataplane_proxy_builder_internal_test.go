package sync

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

// resolveVIPOutbounds used to unconditionally null out dataplane.Spec.Networking.Outbound
// on the VIP-outbounds branch. The Dataplane passed in comes straight from the shared,
// cached xDS mesh context (meshContext.DataplanesByName), so that write was visible to every
// other reader of the same object - the API server, KDS, the mesh-context hasher, etc. - not
// just to this proxy build. See https://github.com/kumahq/kuma/issues/18015 (finding #1).
var _ = Describe("DataplaneProxyBuilder resolveVIPOutbounds", func() {
	It("does not mutate the shared Dataplane's legacy outbound list", func() {
		// given a Dataplane with a legacy outbound, as it would come out of the store
		dataplane := builders.Dataplane().
			WithName("dp-1").
			WithMesh("default").
			AddOutboundToService("backend").
			Build()
		originalOutbounds := dataplane.Spec.GetNetworking().GetOutbound()
		Expect(originalOutbounds).To(HaveLen(1))

		builder := &DataplaneProxyBuilder{}

		// when resolving VIP outbounds (bindOutbounds branch, the one that used to null the field)
		builder.resolveVIPOutbounds(xds_context.MeshContext{}, dataplane, false, true)

		// then the shared Dataplane's legacy outbound list is untouched
		Expect(dataplane.Spec.GetNetworking().GetOutbound()).To(Equal(originalOutbounds))
	})
})
