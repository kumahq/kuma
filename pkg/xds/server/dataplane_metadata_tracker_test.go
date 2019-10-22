package server_test

import (
	"github.com/Kong/kuma/pkg/core/xds"
	"github.com/Kong/kuma/pkg/xds/server"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataplane Metadata Tracker", func() {

	tracker := server.NewDataplaneMetadataTracker()

	req := v2.DiscoveryRequest{
		Node: &core.Node{
			Id: "default.example.pilot",
			Metadata: &types.Struct{
				Fields: map[string]*types.Value{
					"dataplaneTokenPath": &types.Value{
						Kind: &types.Value_StringValue{
							StringValue: "/tmp/token",
						},
					},
				},
			},
		},
	}
	const streamId = 123

	It("should track metadata", func() {
		// when
		err := tracker.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.DataplaneTokenPath).To(Equal("/tmp/token"))

		// when
		tracker.OnStreamClosed(streamId)

		// then metadata should be deleted
		metadata = tracker.Metadata(streamId)
		Expect(metadata).To(Equal(&xds.DataplaneMetadata{}))
	})

	It("should track metadata with empty Node in consecutive DiscoveryRequests", func() {
		// when
		err := tracker.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = tracker.OnStreamRequest(streamId, &v2.DiscoveryRequest{})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.DataplaneTokenPath).To(Equal("/tmp/token"))
	})
})
