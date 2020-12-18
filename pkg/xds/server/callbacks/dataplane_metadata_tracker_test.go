package callbacks_test

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("Dataplane Metadata Tracker", func() {

	tracker := NewDataplaneMetadataTracker()
	callbacks := util_xds_v2.AdaptCallbacks(tracker)

	req := v2.DiscoveryRequest{
		Node: &envoy_core.Node{
			Id: "default.example",
			Metadata: &pstruct.Struct{
				Fields: map[string]*pstruct.Value{
					"dataplaneTokenPath": &pstruct.Value{
						Kind: &pstruct.Value_StringValue{
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
		err := callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.GetDataplaneTokenPath()).To(Equal("/tmp/token"))

		// when
		tracker.OnStreamClosed(streamId)

		// then metadata should be deleted
		metadata = tracker.Metadata(streamId)
		Expect(metadata).To(Equal(&xds.DataplaneMetadata{}))
	})

	It("should track metadata with empty Node in consecutive DiscoveryRequests", func() {
		// when
		err := callbacks.OnStreamRequest(streamId, &req)

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = callbacks.OnStreamRequest(streamId, &v2.DiscoveryRequest{})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(streamId)

		// then
		Expect(metadata.GetDataplaneTokenPath()).To(Equal("/tmp/token"))
	})
})
