package callbacks_test

import (
	"context"
	"time"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("DataplaneStatusTracker", func() {

	var t0 time.Time

	var tracker DataplaneStatusTracker
	var callbacks envoy_server.Callbacks

	var runtimeInfo = test_runtime.TestRuntimeInfo{InstanceId: "test"}
	var ctx context.Context

	// overridden package variables
	var backupNow func() time.Time
	var backupNewUUID func() string

	BeforeEach(func() {
		backupNow = core.Now
		backupNewUUID = core.NewUUID
	})

	AfterEach(func() {
		core.Now = backupNow
		core.NewUUID = backupNewUUID
	})

	BeforeEach(func() {
		t0, _ = time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")
		core.Now = func() time.Time {
			defer func() { t0 = t0.Add(1 * time.Second) }()
			return t0
		}
		core.NewUUID = func() string {
			return "a9680ef2-aa57-11e9-85b6-acde48001122"
		}
	})

	BeforeEach(func() {
		tracker = NewDataplaneStatusTracker(&runtimeInfo, func(dataplaneType core_model.ResourceType, accessor SubscriptionStatusAccessor) DataplaneInsightSink {
			return DataplaneInsightSinkFunc(func(<-chan struct{}) {})
		})
		callbacks = v2.AdaptCallbacks(tracker)
		ctx = context.Background()
	})

	It("should properly handle ADS connection open/close", func() {
		// given
		streamID := int64(1)

		By("simulating start of ADS subscription")
		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")
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
            version:
              kumaDp: {}
              envoy: {}
`))

		By("simulating end of ADS subscription")
		// when
		callbacks.OnStreamClosed(streamID)

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
            version:
              kumaDp: {}
              envoy: {}
`))
	})

	It("should tolerate xDS requests with empty Node", func() {
		// given
		streamID := int64(1)

		By("simulating start of ADS subscription")
		// when
		err := callbacks.OnStreamOpen(ctx, streamID, "")
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
		err = callbacks.OnStreamRequest(streamID, discoveryRequest)
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
        version:
          kumaDp: {}
          envoy: {}
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
			version := util_proto.MustToStruct(&mesh_proto.Version{
				KumaDp: &mesh_proto.KumaDpVersion{
					Version:   "0.0.1",
					GitTag:    "v0.0.1",
					GitCommit: "91ce236824a9d875601679aa80c63783fb0e8725",
					BuildDate: "2019-08-07T11:26:06Z",
				},
				Envoy: &mesh_proto.EnvoyVersion{
					Version: "1.15.0",
					Build:   "hash/1.15.0/RELEASE",
				},
			})

			By("simulating start of subscription")
			// when
			err := callbacks.OnStreamOpen(ctx, streamID, "")
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
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"dataplane.token": {
								Kind: &structpb.Value_StringValue{
									StringValue: "token",
								},
							},
							"version": {
								Kind: &structpb.Value_StructValue{
									StructValue: version,
								},
							},
						},
					},
				},
				TypeUrl: given.TypeUrl,
			}
			err = callbacks.OnStreamRequest(streamID, discoveryRequest)
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
        version:
          kumaDp:
            buildDate: "2019-08-07T11:26:06Z"
            gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
            gitTag: v0.0.1
            version: 0.0.1
          envoy:
            build: hash/1.15.0/RELEASE
            version: 1.15.0
`))

			By("simulating initial xDS response")
			// when
			discoveryResponse := &envoy.DiscoveryResponse{
				TypeUrl: given.TypeUrl,
				Nonce:   "1",
			}
			callbacks.OnStreamResponse(streamID, discoveryRequest, discoveryResponse)
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
			err = callbacks.OnStreamRequest(streamID, discoveryRequest)
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
			err = callbacks.OnStreamRequest(streamID, discoveryRequest)
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
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
            version:
              kumaDp:
                buildDate: "2019-08-07T11:26:06Z"
                gitCommit: 91ce236824a9d875601679aa80c63783fb0e8725
                gitTag: v0.0.1
                version: 0.0.1
              envoy:
                build: hash/1.15.0/RELEASE
                version: 1.15.0
`,
		}),
	)

	type versionTestCase struct {
		version *structpb.Value
	}

	DescribeTable("should read node.metadata without error",
		func(given versionTestCase) {
			// given
			streamID := int64(1)
			discoveryRequest := &envoy.DiscoveryRequest{
				TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
				Node: &envoy_core.Node{
					Id: "default.example-001",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"version": given.version,
						},
					},
				},
			}

			// when
			err := callbacks.OnStreamOpen(context.Background(), streamID, "")
			Expect(err).ToNot(HaveOccurred())
			err = callbacks.OnStreamRequest(streamID, discoveryRequest)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			accessor, _ := tracker.GetStatusAccessor(streamID)
			_, sub := accessor.GetStatus()
			Expect(sub.GetVersion()).ToNot(BeNil())
			Expect(sub.GetVersion()).To(MatchProto(mesh_proto.NewVersion()))
		},
		Entry("when version is a nil struct", versionTestCase{
			version: &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: nil,
				},
			},
		}),
		Entry("when version is not a struct", versionTestCase{
			version: &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: "v1.0.0",
				},
			},
		}),
	)
})

type DataplaneInsightSinkFunc func(stop <-chan struct{})

func (f DataplaneInsightSinkFunc) Start(stop <-chan struct{}) {
	f(stop)
}
