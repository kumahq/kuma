package server

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	status "google.golang.org/genproto/googleapis/rpc/status"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	util_proto "github.com/Kong/kuma/pkg/util/proto"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	test_runtime "github.com/Kong/kuma/pkg/test/runtime"
)

var _ = Describe("DataplaneStatusTracker", func() {

	var t0 time.Time

	var tracker DataplaneStatusTracker
	var runtimeInfo = test_runtime.TestRuntimeInfo{InstanceId: "test"}
	var ctx context.Context

	// overridden package variables
	var backupNow func() time.Time
	var backupNewUUID func() string

	BeforeEach(func() {
		backupNow = now
		backupNewUUID = newUUID
	})

	AfterEach(func() {
		now = backupNow
		newUUID = backupNewUUID
	})

	BeforeEach(func() {
		t0, _ = time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")
		now = func() time.Time {
			defer func() { t0 = t0.Add(1 * time.Second) }()
			return t0
		}
		newUUID = func() string {
			return "a9680ef2-aa57-11e9-85b6-acde48001122"
		}
	})

	BeforeEach(func() {
		tracker = NewDataplaneStatusTracker(runtimeInfo, func(accessor SubscriptionStatusAccessor) DataplaneInsightSink {
			return DataplaneInsightSinkFunc(func(<-chan struct{}) {})
		})
		ctx = context.Background()
	})

	It("should properly handle ADS connection open/close", func() {
		// given
		streamID := int64(1)

		By("simulating start of ADS subscription")
		// when
		err := tracker.OnStreamOpen(ctx, streamID, "")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		accessor, _ := tracker.GetStatusAccessor(streamID)
		// then
		Expect(accessor).ToNot(BeNil())

		By("ensuring ADS subscription initial state")
		// when
		key, subscription := accessor.GetStatus()
		// then
		Expect(key).To(Equal(core_model.ResourceKey{}))
		Expect(util_proto.ToYAML(subscription)).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lds: {}
              rds: {}
              total: {}
`))

		By("simulating end of ADS subscription")
		// when
		tracker.OnStreamClosed(streamID)
		// then

		By("ensuring ADS subscription final state")
		// when
		key, subscription = accessor.GetStatus()
		// then
		Expect(key).To(Equal(core_model.ResourceKey{}))
		Expect(util_proto.ToYAML(subscription)).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            disconnectTime: "2019-07-01T00:00:01Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lds: {}
              rds: {}
              total: {}
`))
	})

	It("should tolerate xDS requests with empty Node", func() {
		// given
		streamID := int64(1)

		By("simulating start of ADS subscription")
		// when
		err := tracker.OnStreamOpen(ctx, streamID, "")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		accessor, _ := tracker.GetStatusAccessor(streamID)
		// then
		Expect(accessor).ToNot(BeNil())

		By("simulating initial LDS request")
		// when
		discoveryRequest := &envoy.DiscoveryRequest{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
		}
		err = tracker.OnStreamRequest(streamID, discoveryRequest)
		// then
		Expect(err).ToNot(HaveOccurred())

		By("ensuring that initial LDS request does not increment stats")
		// when
		key, subscription := accessor.GetStatus()
		// then
		Expect(key).To(Equal(core_model.ResourceKey{}))
		Expect(util_proto.ToYAML(subscription)).To(MatchYAML(`
        connectTime: "2019-07-01T00:00:00Z"
        controlPlaneInstanceId: test
        id: a9680ef2-aa57-11e9-85b6-acde48001122
        status:
          cds: {}
          eds: {}
          lds: {}
          rds: {}
          total: {}
`))
	})

	type testCase struct {
		TypeUrl                    string
		ExpectedStatsAfterResponse string
		ExpectedStatsAfterACK      string
		ExpectedStatsAfterNACK     string
	}

	DescribeTable("should properly handle xDS flow",
		func(given testCase) {
			// given
			streamID := int64(1)

			By("simulating start of subscription")
			// when
			err := tracker.OnStreamOpen(ctx, streamID, "")
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			accessor, _ := tracker.GetStatusAccessor(streamID)
			// then
			Expect(accessor).ToNot(BeNil())

			By("simulating initial xDS request")
			// when
			discoveryRequest := &envoy.DiscoveryRequest{
				Node: &envoy_core.Node{
					Id: "default.example-001",
				},
				TypeUrl: given.TypeUrl,
			}
			err = tracker.OnStreamRequest(streamID, discoveryRequest)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("ensuring that initial xDS request does not increment stats")
			// when
			key, subscription := accessor.GetStatus()
			// then
			Expect(key).To(Equal(core_model.ResourceKey{
				Mesh: "default",
				Name: "example-001",
			}))
			Expect(util_proto.ToYAML(subscription)).To(MatchYAML(`
        connectTime: "2019-07-01T00:00:00Z"
        controlPlaneInstanceId: test
        id: a9680ef2-aa57-11e9-85b6-acde48001122
        status:
          cds: {}
          eds: {}
          lds: {}
          rds: {}
          total: {}
`))

			By("simulating initial xDS response")
			// when
			discoveryResponse := &envoy.DiscoveryResponse{
				TypeUrl: given.TypeUrl,
				Nonce:   "1",
			}
			tracker.OnStreamResponse(streamID, discoveryRequest, discoveryResponse)
			// and
			key, subscription = accessor.GetStatus()
			// then
			Expect(key).To(Equal(core_model.ResourceKey{
				Mesh: "default",
				Name: "example-001",
			}))
			Expect(util_proto.ToYAML(subscription)).To(MatchYAML(given.ExpectedStatsAfterResponse))

			By("simulating xDS ACK request")
			// when
			discoveryRequest = &envoy.DiscoveryRequest{
				TypeUrl:       given.TypeUrl,
				ResponseNonce: "1",
			}
			err = tracker.OnStreamRequest(streamID, discoveryRequest)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("ensuring that xDS ACK request does increment stats")
			// when
			key, subscription = accessor.GetStatus()
			// then
			Expect(key).To(Equal(core_model.ResourceKey{
				Mesh: "default",
				Name: "example-001",
			}))
			Expect(util_proto.ToYAML(subscription)).To(MatchYAML(given.ExpectedStatsAfterACK))

			By("simulating xDS NACK request")
			// when
			discoveryRequest = &envoy.DiscoveryRequest{
				TypeUrl:       given.TypeUrl,
				ResponseNonce: "1",
				ErrorDetail: &status.Status{
					Message: "failed to apply LDS response",
				},
			}
			err = tracker.OnStreamRequest(streamID, discoveryRequest)
			// then
			Expect(err).ToNot(HaveOccurred())

			By("ensuring that xDS NACK request does increment stats")
			// when
			key, subscription = accessor.GetStatus()
			// then
			Expect(key).To(Equal(core_model.ResourceKey{
				Mesh: "default",
				Name: "example-001",
			}))
			Expect(util_proto.ToYAML(subscription)).To(MatchYAML(given.ExpectedStatsAfterNACK))
		},
		Entry("should properly handle LDS flow", testCase{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
			ExpectedStatsAfterResponse: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:01Z"
              lds:
                responsesSent: "1"
              rds: {}
              total:
                responsesSent: "1"
`,
			ExpectedStatsAfterACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:02Z"
              lds:
                responsesAcknowledged: "1"
                responsesSent: "1"
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesSent: "1"
`,
			ExpectedStatsAfterNACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:03Z"
              lds:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
`,
		}),
		Entry("should properly handle RDS flow", testCase{
			TypeUrl: "type.googleapis.com/envoy.api.v2.RouteConfiguration",
			ExpectedStatsAfterResponse: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:01Z"
              lds: {}
              rds:
                responsesSent: "1"
              total:
                responsesSent: "1"
`,
			ExpectedStatsAfterACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:02Z"
              lds: {}
              rds:
                responsesAcknowledged: "1"
                responsesSent: "1"
              total:
                responsesAcknowledged: "1"
                responsesSent: "1"
`,
			ExpectedStatsAfterNACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:03Z"
              lds: {}
              rds:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
              total:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
`,
		}),
		Entry("should properly handle CDS flow", testCase{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
			ExpectedStatsAfterResponse: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds:
                responsesSent: "1"
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:01Z"
              lds: {}
              rds: {}
              total:
                responsesSent: "1"
`,
			ExpectedStatsAfterACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds:
                responsesAcknowledged: "1"
                responsesSent: "1"
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:02Z"
              lds: {}
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesSent: "1"
`,
			ExpectedStatsAfterNACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:03Z"
              lds: {}
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
`,
		}),
		Entry("should properly handle EDS flow", testCase{
			TypeUrl: "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
			ExpectedStatsAfterResponse: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds:
                responsesSent: "1"
              lastUpdateTime: "2019-07-01T00:00:01Z"
              lds: {}
              rds: {}
              total:
                responsesSent: "1"
`,
			ExpectedStatsAfterACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds:
                responsesAcknowledged: "1"
                responsesSent: "1"
              lastUpdateTime: "2019-07-01T00:00:02Z"
              lds: {}
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesSent: "1"
`,
			ExpectedStatsAfterNACK: `
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: test
            id: a9680ef2-aa57-11e9-85b6-acde48001122
            status:
              cds: {}
              eds:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
              lastUpdateTime: "2019-07-01T00:00:03Z"
              lds: {}
              rds: {}
              total:
                responsesAcknowledged: "1"
                responsesRejected: "1"
                responsesSent: "1"
`,
		}),
	)
})

type DataplaneInsightSinkFunc func(stop <-chan struct{})

func (f DataplaneInsightSinkFunc) Start(stop <-chan struct{}) {
	f(stop)
}
