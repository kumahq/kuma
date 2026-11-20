package server_test

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/kds/server"
	"github.com/kumahq/kuma/v3/pkg/kds/util"
)

var _ = Describe("Hasher", func() {
	DescribeTable("should build the snapshot cache key",
		func(tenant string, streamID int64, expected string) {
			// given
			node := &envoy_core.Node{Id: "zone-1"}
			if tenant != "" {
				util.FillTenantMetadata(tenant, node)
			}
			if streamID != 0 {
				util.FillStreamMetadata(streamID, node)
			}

			// when
			id := server.Hasher{}.ID(node)

			// then
			Expect(id).To(Equal(expected))
		},
		Entry("without metadata", "", int64(0), "zone-1"),
		Entry("with tenant", "tenant-a", int64(0), "zone-1:tenant-a"),
		Entry("with stream", "", int64(7), "zone-1/7"),
		Entry("with tenant and stream", "tenant-a", int64(7), "zone-1:tenant-a/7"),
	)

	It("should not depend on the order the metadata was filled in", func() {
		// given
		tenantFirst := &envoy_core.Node{Id: "zone-1"}
		util.FillTenantMetadata("tenant-a", tenantFirst)
		util.FillStreamMetadata(7, tenantFirst)
		streamFirst := &envoy_core.Node{Id: "zone-1"}
		util.FillStreamMetadata(7, streamFirst)
		util.FillTenantMetadata("tenant-a", streamFirst)

		// expect
		Expect(server.Hasher{}.ID(tenantFirst)).To(Equal(server.Hasher{}.ID(streamFirst)))
	})
})
