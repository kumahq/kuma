package callbacks_test

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("Dataplane Metadata Tracker", func() {

	tracker := NewDataplaneMetadataTracker()
	callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

	dpKey := core_model.ResourceKey{
		Mesh: "default",
		Name: "example",
	}
	req := envoy_sd.DiscoveryRequest{
		Node: &envoy_core.Node{
			Id: "default.example",
			Metadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"dataplane.dns.port": {
						Kind: &structpb.Value_StringValue{
							StringValue: "9090",
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
		Expect(metadata.GetDNSPort()).To(Equal(uint32(9090)))

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
		err = callbacks.OnStreamRequest(streamId, &envoy_sd.DiscoveryRequest{})

		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		metadata := tracker.Metadata(dpKey)

		// then
		Expect(metadata.GetDNSPort()).To(Equal(uint32(9090)))
	})
})
