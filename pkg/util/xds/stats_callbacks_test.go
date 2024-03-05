package xds_test

import (
	"context"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/kumahq/kuma/pkg/core"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var _ = Describe("Stats callbacks", func() {
	Context("Callbacks", func() {
		const streamId = int64(1)
		var adaptedCallbacks envoy_xds.Callbacks
		var statsCallbacks util_xds.StatsCallbacks
		var metrics core_metrics.Metrics
		var currentTime time.Time

		BeforeEach(func() {
			m, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			metrics = m
			statsCallbacks, err = util_xds.NewStatsCallbacks(metrics, "xds")
			adaptedCallbacks = util_xds_v3.AdaptCallbacks(statsCallbacks)
			Expect(err).ToNot(HaveOccurred())

			currentTime = time.Now()
			core.Now = func() time.Time {
				return currentTime
			}
		})

		AfterEach(func() {
			core.Now = time.Now
		})

		It("should track active streams", func() {
			// when
			err := statsCallbacks.OnStreamOpen(context.Background(), streamId, "")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_streams_active").GetGauge().GetValue()).To(Equal(1.0))

			// when
			statsCallbacks.OnStreamClosed(streamId)
			Expect(test_metrics.FindMetric(metrics, "xds_streams_active").GetGauge().GetValue()).To(Equal(0.0))
		})

		It("should ignore initial DiscoveryRequest", func() {
			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "",
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_requests_received").GetCounter().GetValue()).To(Equal(0.0))
		})

		It("should track ACK", func() {
			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_requests_received", "confirmation", "ACK", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track NACK", func() {
			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
				ErrorDetail: &status.Status{},
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_requests_received", "confirmation", "NACK", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track responses sent", func() {
			// when
			resp := &envoy_discovery.DiscoveryResponse{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
			}
			adaptedCallbacks.OnStreamResponse(context.Background(), streamId, nil, resp)

			// then
			Expect(test_metrics.FindMetric(metrics, "xds_responses_sent", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track config delivery", func() {
			// given
			statsCallbacks.ConfigReadyForDelivery("123")
			currentTime = currentTime.Add(time.Second)

			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_delivery").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
			Expect(test_metrics.FindMetric(metrics, "xds_delivery").GetSummary().GetSampleSum()).To(Equal(float64(time.Second.Milliseconds())))
		})

		It("should not track delivery of configs that were not ready to being delivered", func() {
			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_delivery").GetSummary().GetSampleCount()).To(Equal(uint64(0)))
		})

		It("should not track discarded configs", func() {
			// given
			statsCallbacks.ConfigReadyForDelivery("123")
			statsCallbacks.DiscardConfig("123")

			// when
			req := &envoy_discovery.DiscoveryRequest{
				TypeUrl:     resource.RouteType,
				VersionInfo: "123",
			}
			err := adaptedCallbacks.OnStreamRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "xds_delivery").GetSummary().GetSampleCount()).To(Equal(uint64(0)))
		})
	})

	Context("DeltaCallbacks", func() {
		const streamId = int64(1)
		var adaptedCallbacks envoy_xds.Callbacks
		var statsCallbacks util_xds.StatsCallbacks
		var metrics core_metrics.Metrics
		var currentTime time.Time

		BeforeEach(func() {
			m, err := core_metrics.NewMetrics("Zone")
			Expect(err).ToNot(HaveOccurred())
			metrics = m
			statsCallbacks, err = util_xds.NewStatsCallbacks(metrics, "delta_xds")
			adaptedCallbacks = util_xds_v3.AdaptDeltaCallbacks(statsCallbacks)
			Expect(err).ToNot(HaveOccurred())

			currentTime = time.Now()
			core.Now = func() time.Time {
				return currentTime
			}
		})

		AfterEach(func() {
			core.Now = time.Now
		})

		It("should track active streams", func() {
			// when
			err := statsCallbacks.OnDeltaStreamOpen(context.Background(), streamId, "")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "delta_xds_streams_active").GetGauge().GetValue()).To(Equal(1.0))

			// when
			statsCallbacks.OnStreamClosed(streamId)
			Expect(test_metrics.FindMetric(metrics, "delta_xds_streams_active").GetGauge().GetValue()).To(Equal(0.0))
		})

		It("should ignore initial DiscoveryRequest", func() {
			// when
			req := &envoy_discovery.DeltaDiscoveryRequest{
				TypeUrl: resource.RouteType,
			}
			err := adaptedCallbacks.OnStreamDeltaRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "delta_xds_requests_received").GetCounter().GetValue()).To(Equal(0.0))
		})

		It("should track ACK", func() {
			// when
			req := &envoy_discovery.DeltaDiscoveryRequest{
				TypeUrl: resource.RouteType,
				InitialResourceVersions: map[string]string{
					"route-1": "123",
				},
				ResponseNonce: "1",
			}
			err := adaptedCallbacks.OnStreamDeltaRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "delta_xds_requests_received", "confirmation", "ACK", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track NACK", func() {
			// when
			req := &envoy_discovery.DeltaDiscoveryRequest{
				TypeUrl: resource.RouteType,
				InitialResourceVersions: map[string]string{
					"route-1": "123",
				},
				ResponseNonce: "1",
				ErrorDetail:   &status.Status{},
			}
			err := adaptedCallbacks.OnStreamDeltaRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "delta_xds_requests_received", "confirmation", "NACK", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track responses sent", func() {
			// when
			resp := &envoy_discovery.DeltaDiscoveryResponse{
				TypeUrl:           resource.RouteType,
				SystemVersionInfo: "123",
			}
			adaptedCallbacks.OnStreamDeltaResponse(streamId, nil, resp)

			// then
			Expect(test_metrics.FindMetric(metrics, "delta_xds_responses_sent", "type_url", resource.RouteType).GetCounter().GetValue()).To(Equal(1.0))
		})

		It("should track config delivery", func() {
			// given
			nodeID := "x"
			statsCallbacks.ConfigReadyForDelivery(nodeID + resource.RouteType)
			currentTime = currentTime.Add(time.Second)

			// when
			req := &envoy_discovery.DeltaDiscoveryRequest{
				Node: &corev3.Node{
					Id: nodeID,
				},
				TypeUrl:       resource.RouteType,
				ResponseNonce: "1",
				ErrorDetail:   &status.Status{},
			}
			err := adaptedCallbacks.OnStreamDeltaRequest(streamId, req)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(test_metrics.FindMetric(metrics, "delta_xds_delivery").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
			Expect(test_metrics.FindMetric(metrics, "delta_xds_delivery").GetSummary().GetSampleSum()).To(Equal(float64(time.Second.Milliseconds())))
		})
	})
})
