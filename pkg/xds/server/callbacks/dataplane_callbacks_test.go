package callbacks_test

import (
	"context"
	"time"

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
	OnProxyDisconnectedCounter int
}

func (c *countingDpCallbacks) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.OnProxyConnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.OnProxyDisconnectedCounter++
}

var _ DataplaneCallbacks = &countingDpCallbacks{}

var _ = Describe("Dataplane Callbacks", func() {
	countingCallbacks := &countingDpCallbacks{}
	callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks))

	node := &envoy_core.Node{
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
	}
	req := envoy_sd.DiscoveryRequest{
		Node: node,
	}

	It("should call DataplaneCallbacks correctly", func() {
		// when only OnStreamOpen is called
		err := callbacks.OnStreamOpen(context.Background(), 1, "")
		Expect(err).ToNot(HaveOccurred())

		// then OnProxyConnected and OnProxyReconnected is not yet called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(0))

		// when OnStreamRequest is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then only OnProxyConnected should be called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))

		// when next OnStreamRequest on the same stream is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then OnProxyReconnected and OnProxyReconnected are not called again, they should be only called on the first DiscoveryRequest
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))

		// when next stream for given data plane proxy is connected
		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("already an active stream"))

		// then OnProxyReconnected should not be called twice
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))

		// when first stream is closed
		callbacks.OnStreamClosed(1, node)

		// then OnProxyDisconnected should be called
		Expect(countingCallbacks.OnProxyDisconnectedCounter).To(Equal(1))

		// when last stream is closed
		callbacks.OnStreamClosed(2, node)

		// then OnProxyDisconnected should not be called twice
		Expect(countingCallbacks.OnProxyDisconnectedCounter).To(Equal(1))
	})

	It("should reject new stream when the cleanup is not complete", func() {
		// setup
		cleanupCost := 2 * time.Second
		cleanupStarted := make(chan struct{})
		tracker := NewDataplaneSyncTracker(func(key core_model.ResourceKey, stopped func(key core_model.ResourceKey)) util_xds_v3.Watchdog {
			return WatchdogFunc(func(ctx context.Context) {
				<-ctx.Done()
				cleanupStarted <- struct{}{}
				<-time.After(cleanupCost)
				stopped(key)
			})
		})
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(tracker))

		// when one stream for backend-01 is connected and request is sent
		streamID1 := int64(1)
		streamID2 := int64(2)
		streamID3 := int64(3)
		err := callbacks.OnStreamOpen(context.Background(), streamID1, "")
		Expect(err).ToNot(HaveOccurred())
		node := &envoy_core.Node{Id: "default.backend-01"}
		err = callbacks.OnStreamRequest(streamID1, &envoy_sd.DiscoveryRequest{Node: node})
		Expect(err).ToNot(HaveOccurred())

		// and when a new stream from backend-01 tries to connect quickly before the cleanup is complete
		callbacks.OnStreamClosed(streamID1, node)
		<-cleanupStarted
		<-time.After(cleanupCost - time.Second)

		err = callbacks.OnStreamOpen(context.Background(), streamID2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(streamID2, &envoy_sd.DiscoveryRequest{Node: node})

		// then it should be rejected
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("already an active stream"))
		callbacks.OnStreamClosed(streamID2, node)

		// and when a new stream from backend-01 tries to connect quickly after the cleanup is complete
		<-time.After(cleanupCost)
		// and when the third stream from backend-01 is connected after the first active stream closed and request is sent
		err = callbacks.OnStreamOpen(context.Background(), streamID3, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(streamID3, &envoy_sd.DiscoveryRequest{Node: node})
		Expect(err).ToNot(HaveOccurred())

		// when other stream is closed and the third stream is open
		callbacks.OnStreamClosed(streamID3, node)
	})
})
