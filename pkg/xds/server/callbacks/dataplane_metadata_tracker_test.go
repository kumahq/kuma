package callbacks_test

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("Dataplane Metadata Tracker", func() {

	tracker := NewDataplaneMetadataTracker()
	callbacks := util_xds_v2.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

	dpKey := core_model.ResourceKey{
		Mesh: "default",
		Name: "example",
	}
	req := v2.DiscoveryRequest{
		Node: &envoy_core.Node{
			Id: "default.example",
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.token": &structpb.Value{
						Kind: &structpb.Value_StringValue{
							StringValue: "token",
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
		metadata := tracker.Metadata(dpKey)

		// then
		Expect(metadata.GetDataplaneToken()).To(Equal("token"))

		// when
		callbacks.OnStreamClosed(streamId)

		// then metadata should be deleted
		metadata = tracker.Metadata(dpKey)
		Expect(metadata).To(BeNil())
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
		metadata := tracker.Metadata(dpKey)

		// then
		Expect(metadata.GetDataplaneToken()).To(Equal("token"))
	})
})
