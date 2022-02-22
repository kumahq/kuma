package callbacks_test

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	. "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("DataplaneStatusTracker", func() {

	var tracker DataplaneStatusTracker
	var callbacks envoy_server.Callbacks

	var runtimeInfo = test_runtime.TestRuntimeInfo{InstanceId: "test"}
	var ctx context.Context

	BeforeEach(func() {
		tracker = NewDataplaneStatusTracker(&runtimeInfo, func(dataplaneType core_model.ResourceType, accessor SubscriptionStatusAccessor) DataplaneInsightSink {
			return DataplaneInsightSinkFunc(func(<-chan struct{}) {})
		})
		callbacks = v3.AdaptCallbacks(tracker)
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
		Expect(subscription.ConnectTime.GetNanos()).ToNot(BeZero())
		Expect(subscription.DisconnectTime.GetNanos()).To(BeZero())

		By("simulating end of ADS subscription")
		// when
		callbacks.OnStreamClosed(streamID)

		By("ensuring ADS subscription final state")
		// when
		key, subscription = accessor.GetStatus()
		// then
		Expect(key).To(Equal(core_model.ResourceKey{}))
		Expect(subscription.DisconnectTime.AsTime().UnixNano()).To(BeNumerically(">=", subscription.ConnectTime.AsTime().UnixNano()))
	})

	zeroStatus := mesh_proto.DiscoverySubscriptionStatus{
		Total: &mesh_proto.DiscoveryServiceStats{},
		Cds:   &mesh_proto.DiscoveryServiceStats{},
		Eds:   &mesh_proto.DiscoveryServiceStats{},
		Lds:   &mesh_proto.DiscoveryServiceStats{},
		Rds:   &mesh_proto.DiscoveryServiceStats{},
	}

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
		discoveryRequest := &envoy_sd.DiscoveryRequest{
			TypeUrl: "type.googleapis.com/envoy.config.listener.v3.Listener",
		}
		err = callbacks.OnStreamRequest(streamID, discoveryRequest)
		// then
		Expect(err).ToNot(HaveOccurred())

		By("ensuring that initial LDS request does not increment stats")
		// when
		key, subscription := accessor.GetStatus()
		// then
		Expect(key).To(Equal(core_model.ResourceKey{}))
		Expect(subscription.Status).To(MatchProto(&zeroStatus))
	})

	type testCase struct {
		TypeUrl   string
		TypeStats string
	}

	DescribeTable("should properly handle xDS flow",
		func(given testCase) {
			// given
			streamID := int64(1)
			version := mesh_proto.Version{
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
			}

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
			discoveryRequest := &envoy_sd.DiscoveryRequest{
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
									StructValue: util_proto.MustToStruct(&version),
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
			Expect(subscription.Status).To(MatchProto(&zeroStatus))
			Expect(subscription.ConnectTime.AsTime().UnixNano()).NotTo(BeZero())
			Expect(subscription.Version).To(MatchProto(&version))

			By("simulating initial xDS response")
			// when
			discoveryResponse := &envoy_sd.DiscoveryResponse{
				TypeUrl: given.TypeUrl,
				Nonce:   "1",
			}
			callbacks.OnStreamResponse(context.TODO(), streamID, discoveryRequest, discoveryResponse)
			// and
			key, subscription = accessor.GetStatus()
			// then
			Expect(key).To(Equal(core_model.ResourceKey{
				Mesh: "default",
				Name: "example-001",
			}))
			sent := &mesh_proto.DiscoveryServiceStats{
				ResponsesSent: 1,
			}
			sentTime := subscription.Status.LastUpdateTime.AsTime().UnixNano()
			Expect(subscription.Status).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Total":         MatchProto(sent),
				"Cds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Eds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Lds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Rds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				given.TypeStats: MatchProto(sent),
			})))
			Expect(sentTime).NotTo(BeZero())

			By("simulating xDS ACK request")
			// when
			discoveryRequest = &envoy_sd.DiscoveryRequest{
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
			acked := &mesh_proto.DiscoveryServiceStats{
				ResponsesAcknowledged: 1,
				ResponsesSent:         1,
			}
			ackTime := subscription.Status.LastUpdateTime.AsTime().UnixNano()
			Expect(subscription.Status).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Total":         MatchProto(acked),
				"Cds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Eds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Lds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Rds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				given.TypeStats: MatchProto(acked),
			})))
			Expect(ackTime).To(BeNumerically(">", sentTime))

			By("simulating xDS NACK request")
			// when
			discoveryRequest = &envoy_sd.DiscoveryRequest{
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
			nacked := &mesh_proto.DiscoveryServiceStats{
				ResponsesRejected:     1,
				ResponsesAcknowledged: 1,
				ResponsesSent:         1,
			}
			nackTime := subscription.Status.LastUpdateTime.AsTime().UnixNano()
			Expect(subscription.Status).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Total":         MatchProto(nacked),
				"Cds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Eds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Lds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				"Rds":           MatchProto(&mesh_proto.DiscoveryServiceStats{}),
				given.TypeStats: MatchProto(nacked),
			})))
			Expect(nackTime).To(BeNumerically(">", ackTime))
		},
		Entry("should properly handle LDS flow", testCase{
			TypeUrl:   "type.googleapis.com/envoy.config.listener.v3.Listener",
			TypeStats: "Lds",
		}),
		Entry("should properly handle RDS flow", testCase{
			TypeUrl:   "type.googleapis.com/envoy.config.route.v3.RouteConfiguration",
			TypeStats: "Rds",
		}),
		Entry("should properly handle CDS flow", testCase{
			TypeUrl:   "type.googleapis.com/envoy.config.cluster.v3.Cluster",
			TypeStats: "Cds",
		}),
		Entry("should properly handle EDS flow", testCase{
			TypeUrl:   "type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment",
			TypeStats: "Eds",
		}),
	)

	type versionTestCase struct {
		version *structpb.Value
	}

	DescribeTable("should read node.metadata without error",
		func(given versionTestCase) {
			// given
			streamID := int64(1)
			discoveryRequest := &envoy_sd.DiscoveryRequest{
				TypeUrl: "type.googleapis.com/envoy.config.listener.v3.Listener",
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
