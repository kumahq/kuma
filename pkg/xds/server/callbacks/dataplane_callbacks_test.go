package callbacks_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

type countingDpCallbacks struct {
	OnStreamConnectedCounter        int
	OnFirstStreamConnectedCounter   int
	OnStreamDisconnectedCounter     int
	OnLastStreamDisconnectedCounter int
}

func (c *countingDpCallbacks) OnStreamConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.OnStreamConnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnFirstStreamConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.OnFirstStreamConnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnStreamDisconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.OnStreamDisconnectedCounter++
}

func (c *countingDpCallbacks) OnLastStreamDisconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.OnLastStreamDisconnectedCounter++
}

var _ DataplaneCallbacks = &countingDpCallbacks{}

var _ = Describe("Dataplane Callbacks", func() {

	countingCallbacks := &countingDpCallbacks{}
	callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks))

	req := envoy_sd.DiscoveryRequest{
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

	It("should call DataplaneCallbacks correctly", func() {
		// when only OnStreamOpen is called
		err := callbacks.OnStreamOpen(context.Background(), 1, "")
		Expect(err).ToNot(HaveOccurred())

		// then OnStreamConnected and OnFirstStreamConnectedCounter is not yet called
		Expect(countingCallbacks.OnStreamConnectedCounter).To(Equal(0))
		Expect(countingCallbacks.OnFirstStreamConnectedCounter).To(Equal(0))

		// when OnStreamRequest is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then both OnStreamConnected and OnFirstStreamConnected should be called
		Expect(countingCallbacks.OnStreamConnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnFirstStreamConnectedCounter).To(Equal(1))

		// when next OnStreamRequest on the same stream is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then OnStreamConnected and OnFirstStreamConnectedCounter are not called again, they should be only called on the first DiscoveryRequest
		Expect(countingCallbacks.OnStreamConnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnFirstStreamConnectedCounter).To(Equal(1))

		// when next stream for given data plane proxy is connected
		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).ToNot(HaveOccurred())

		// then only OnStreamConnected should be called
		Expect(countingCallbacks.OnStreamConnectedCounter).To(Equal(2))
		Expect(countingCallbacks.OnFirstStreamConnectedCounter).To(Equal(1))

		// when first stream is closed
		callbacks.OnStreamClosed(1)

		// then only OnStreamDisconnected should be called
		Expect(countingCallbacks.OnStreamDisconnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnLastStreamDisconnectedCounter).To(Equal(0))

		// when last stream is closed
		callbacks.OnStreamClosed(2)

		// then both OnStreamDisconnected and OnLastStreamDisconnected are called
		Expect(countingCallbacks.OnStreamDisconnectedCounter).To(Equal(2))
		Expect(countingCallbacks.OnLastStreamDisconnectedCounter).To(Equal(1))
	})
})
