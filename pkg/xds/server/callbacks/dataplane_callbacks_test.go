package callbacks_test

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	util_xds_v3 "github.com/kumahq/kuma/v2/pkg/util/xds/v3"
	xds_metrics "github.com/kumahq/kuma/v2/pkg/xds/metrics"
	. "github.com/kumahq/kuma/v2/pkg/xds/server/callbacks"
)

type countingDpCallbacks struct {
	mu                         sync.Mutex
	OnProxyConnectedCounter    int
	OnProxyDisconnectedCounter int
}

func (c *countingDpCallbacks) OnProxyConnected(streamID core_xds.StreamID, dpKey core_model.ResourceKey, ctx context.Context, metadata core_xds.DataplaneMetadata) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OnProxyConnectedCounter++
	return nil
}

func (c *countingDpCallbacks) OnProxyDisconnected(ctx context.Context, streamID core_xds.StreamID, dpKey core_model.ResourceKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OnProxyDisconnectedCounter++
}

var _ DataplaneCallbacks = &countingDpCallbacks{}

type blockingConnectCallbacks struct {
	mu                      sync.Mutex
	started                 chan struct{}
	release                 chan struct{}
	blockOn                 int
	OnProxyConnectedCounter int
}

func (c *blockingConnectCallbacks) OnProxyConnected(core_xds.StreamID, core_model.ResourceKey, context.Context, core_xds.DataplaneMetadata) error {
	c.mu.Lock()
	c.OnProxyConnectedCounter++
	if c.blockOn == 0 {
		c.blockOn = 1
	}
	if c.OnProxyConnectedCounter == c.blockOn {
		close(c.started)
		c.mu.Unlock()
		<-c.release
		return nil
	}
	c.mu.Unlock()
	return nil
}

func (c *blockingConnectCallbacks) OnProxyDisconnected(context.Context, core_xds.StreamID, core_model.ResourceKey) {
}

var _ DataplaneCallbacks = &blockingConnectCallbacks{}

var _ = Describe("Dataplane Callbacks", func() {
	countingCallbacks := &countingDpCallbacks{}
	callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks, nil))

	node := &envoy_core.Node{
		Id: "default.example",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				core_xds.FieldDataplaneProxyType: {
					Kind: &structpb.Value_StringValue{
						StringValue: string(mesh_proto.DataplaneProxyType),
					},
				},
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

		// then OnProxyConnected is not yet called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(0))

		// when OnStreamRequest is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then only OnProxyConnected should be called
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))

		// when next OnStreamRequest on the same stream is sent
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// then OnProxyReconnected is not called again, they should be only called on the first DiscoveryRequest
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(1))

		// when next stream for given data plane proxy is connected
		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("already an active stream"))

		// then OnProxyConnected should not be called twice
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

	It("should takeover stale active stream", func() {
		countingCallbacks := &countingDpCallbacks{}
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks, nil))

		ctx1, cancel1 := context.WithCancel(context.Background())
		err := callbacks.OnStreamOpen(ctx1, 1, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(1, &req)
		Expect(err).ToNot(HaveOccurred())

		// Mark owner stream as stale before cleanup callback runs.
		cancel1()

		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).ToNot(HaveOccurred())

		callbacks.OnStreamClosed(1, node)
		callbacks.OnStreamClosed(2, node)

		countingCallbacks.mu.Lock()
		defer countingCallbacks.mu.Unlock()
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(2))
		Expect(countingCallbacks.OnProxyDisconnectedCounter).To(Equal(1))
	})

	It("should return resource exhausted while registration in progress", func() {
		blockingCallbacks := &blockingConnectCallbacks{
			started: make(chan struct{}),
			release: make(chan struct{}),
			blockOn: 1,
		}
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(blockingCallbacks, nil))

		err := callbacks.OnStreamOpen(context.Background(), 1, "")
		Expect(err).ToNot(HaveOccurred())

		resultCh := make(chan error, 1)
		go func() {
			resultCh <- callbacks.OnStreamRequest(1, &req)
		}()

		Eventually(blockingCallbacks.started).Should(BeClosed())

		err = callbacks.OnStreamOpen(context.Background(), 2, "")
		Expect(err).ToNot(HaveOccurred())
		err = callbacks.OnStreamRequest(2, &req)
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.ResourceExhausted))

		close(blockingCallbacks.release)
		Expect(<-resultCh).ToNot(HaveOccurred())
	})

	It("should not remove new owner from activeStreams when stale owner stream closes", func() {
		// This test verifies the owner-ID guard in onStreamClosed: when a stale
		// stream is taken over by a new one, the stale owner's OnStreamClosed must
		// not evict the new owner's entry from activeStreams.
		errorCh := make(chan error, 1)
		countingCallbacks := &countingDpCallbacks{}
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks, nil))

		// Stream 1 opens and registers.
		ctx1, cancel1 := context.WithCancel(context.Background())
		Expect(callbacks.OnStreamOpen(ctx1, 1, "")).To(Succeed())
		Expect(callbacks.OnStreamRequest(1, &req)).To(Succeed())

		// Cancel ctx to make stream 1 stale.
		cancel1()

		// Stream 2 registers via stale takeover.
		Expect(callbacks.OnStreamOpen(context.Background(), 2, "")).To(Succeed())
		Expect(callbacks.OnStreamRequest(2, &req)).To(Succeed())

		// Stream 1 (stale owner) closes; the owner-ID guard should prevent it from
		// removing stream 2 from activeStreams.
		callbacks.OnStreamClosed(1, node)

		// Stream 3 should now be rejected because stream 2 is still the active owner.
		Expect(callbacks.OnStreamOpen(context.Background(), 3, "")).To(Succeed())
		go func() {
			errorCh <- callbacks.OnStreamRequest(3, &req)
		}()
		Eventually(errorCh).Should(Receive(MatchError(ContainSubstring("already an active stream"))))

		// Clean up stream 2.
		callbacks.OnStreamClosed(2, node)
		callbacks.OnStreamClosed(3, node)

		countingCallbacks.mu.Lock()
		defer countingCallbacks.mu.Unlock()
		Expect(countingCallbacks.OnProxyConnectedCounter).To(Equal(2))
	})

	It("should allow only one winner in concurrent stale takeover race", func() {
		countingCallbacks := &countingDpCallbacks{}
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(countingCallbacks, nil))

		ctx1, cancel1 := context.WithCancel(context.Background())
		Expect(callbacks.OnStreamOpen(ctx1, 1, "")).To(Succeed())
		Expect(callbacks.OnStreamRequest(1, &req)).To(Succeed())

		// Mark stream 1 as stale before cleanup callback runs.
		cancel1()

		Expect(callbacks.OnStreamOpen(context.Background(), 2, "")).To(Succeed())
		Expect(callbacks.OnStreamOpen(context.Background(), 3, "")).To(Succeed())

		results := make(chan error, 2)
		go func() { results <- callbacks.OnStreamRequest(2, &req) }()
		go func() { results <- callbacks.OnStreamRequest(3, &req) }()

		err1 := <-results
		err2 := <-results

		successes := 0
		failures := 0
		for _, err := range []error{err1, err2} {
			if err == nil {
				successes++
				continue
			}
			failures++
			Expect(
				status.Code(err) == codes.ResourceExhausted ||
					err.Error() == "there is already an active stream from this node, try again later",
			).To(BeTrue())
		}

		Expect(successes).To(Equal(1))
		Expect(failures).To(Equal(1))

		callbacks.OnStreamClosed(1, node)
		callbacks.OnStreamClosed(2, node)
		callbacks.OnStreamClosed(3, node)
	})

	It("should increment stale takeover and in-progress retry counters", func() {
		baseMetrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		xdsMetrics, err := xds_metrics.NewMetrics(baseMetrics)
		Expect(err).ToNot(HaveOccurred())

		blocker := &blockingConnectCallbacks{
			started: make(chan struct{}),
			release: make(chan struct{}),
			blockOn: 2,
		}
		callbacks := util_xds_v3.AdaptCallbacks(DataplaneCallbacksToXdsCallbacks(blocker, xdsMetrics))

		ctx1, cancel1 := context.WithCancel(context.Background())
		Expect(callbacks.OnStreamOpen(ctx1, 1, "")).To(Succeed())
		Expect(callbacks.OnStreamRequest(1, &req)).To(Succeed())
		cancel1()

		// stale owner takeover path
		Expect(callbacks.OnStreamOpen(context.Background(), 2, "")).To(Succeed())
		resultCh := make(chan error, 1)
		go func() {
			resultCh <- callbacks.OnStreamRequest(2, &req)
		}()

		Eventually(blocker.started).Should(BeClosed())

		// connecting in-progress retry path
		Expect(callbacks.OnStreamOpen(context.Background(), 3, "")).To(Succeed())
		err = callbacks.OnStreamRequest(3, &req)
		Expect(err).To(HaveOccurred())
		Expect(status.Code(err)).To(Equal(codes.ResourceExhausted))

		close(blocker.release)
		Expect(<-resultCh).ToNot(HaveOccurred())

		inProgressCounter, err := readCounter(xdsMetrics.XdsStreamRegistrationInProgressRetries.WithLabelValues("default", string(mesh_proto.DataplaneProxyType)))
		Expect(err).ToNot(HaveOccurred())
		staleCounter, err := readCounter(xdsMetrics.XdsStreamStaleOwnerTakeovers.WithLabelValues("default", string(mesh_proto.DataplaneProxyType)))
		Expect(err).ToNot(HaveOccurred())
		Expect(staleCounter).To(Equal(float64(1)))
		Expect(inProgressCounter).To(Equal(float64(1)))

		callbacks.OnStreamClosed(1, node)
		callbacks.OnStreamClosed(2, node)
		callbacks.OnStreamClosed(3, node)
	})
})

func readCounter(counter interface{ Write(*dto.Metric) error }) (float64, error) {
	metric := &dto.Metric{}
	if err := counter.Write(metric); err != nil {
		return 0, err
	}
	return metric.GetCounter().GetValue(), nil
}
