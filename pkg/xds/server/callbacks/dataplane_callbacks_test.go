package callbacks_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

type countingDpCallbacks struct {
	OnProxyConnectedCounter    int
	OnProxyReconnectedCounter  int
	OnProxyDisconnectedCounter int
}

func (c *countingDpCallbacks) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.OnProxyConnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnProxyReconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.OnProxyReconnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnProxyDisconnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.OnProxyDisconnectedCounter++
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
					"dataplane.token": {
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

		// then OnProxyConnected and OnProxyReconnected is not yet called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(0))
		Expect(countingCallbacks.OnProxyReconnectedCounter).To(Equal(0))

		// when OnStreamRequest is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then only OnProxyConnected should be called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnProxyReconnectedCounter).To(Equal(0))

		// when next OnStreamRequest on the same stream is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then OnProxyReconnected and OnProxyReconnected are not called again, they should be only called on the first DiscoveryRequest
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnProxyReconnectedCounter).To(Equal(0))

		// when next stream for given data plane proxy is connected
		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).ToNot(HaveOccurred())

		// then only OnProxyReconnected should be called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))
		Expect(countingCallbacks.OnProxyReconnectedCounter).To(Equal(1))

		// when first stream is closed
		callbacks.OnStreamClosed(1)

		// then OnProxyDisconnected should not yet be called
		Expect(countingCallbacks.OnProxyDisconnectedCounter).To(Equal(0))

		// when last stream is closed
		callbacks.OnStreamClosed(2)

		// then OnProxyDisconnected should be called
		Expect(countingCallbacks.OnProxyDisconnectedCounter).To(Equal(1))
	})
})
