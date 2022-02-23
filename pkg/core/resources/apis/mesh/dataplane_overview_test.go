package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("DataplaneOverview", func() {

	Describe("NewDataplaneOverviews", func() {
		It("should create overviews from dataplanes and insights", func() {
			dataplanes := DataplaneResourceList{Items: []*DataplaneResource{
				{
					Meta: &model.ResourceMeta{
						Name: "dp1",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.Dataplane{},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "dp2",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.Dataplane{},
				},
			}}

			insights := DataplaneInsightResourceList{Items: []*DataplaneInsightResource{
				{
					Meta: &model.ResourceMeta{
						Name: "dp1",
						Mesh: "mesh1",
					},
					Spec: &mesh_proto.DataplaneInsight{},
				},
			}}

			overviews := NewDataplaneOverviews(dataplanes, insights)
			Expect(overviews.Items).To(HaveLen(2))
			Expect(overviews.Items[0].Spec.Dataplane).To(MatchProto(dataplanes.Items[0].Spec))
			Expect(overviews.Items[0].Spec.DataplaneInsight).To(MatchProto(insights.Items[0].Spec))
			Expect(overviews.Items[1].Spec.Dataplane).To(MatchProto(dataplanes.Items[1].Spec))
			Expect(overviews.Items[1].Spec.DataplaneInsight).To(BeNil())
		})
	})

	Context("GetStatus", func() {
		type testCase struct {
			overview   *mesh_proto.DataplaneOverview
			status     Status
			errReasons []string
		}

		DescribeTable("should compute status",
			func(given testCase) {
				// given
				resource := NewDataplaneOverviewResource()
				resource.Spec = given.overview

				// when
				status, errReasons := resource.GetStatus()

				// then
				Expect(status).To(Equal(given.status))
				Expect(errReasons).To(Equal(given.errReasons))
			},
			Entry("online when proxy is connected and health is nil", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Health: nil,
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Online,
			}),
			Entry("online when proxy is connected and health is ready", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
										Ready: true,
									},
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Online,
			}),
			Entry("online when proxy is connected and is gateway", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Online,
			}),
			Entry("offline when proxy is not connected even that inbound is ready", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
										Ready: true,
									},
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime:    proto.MustTimestampProto(core.Now()),
								DisconnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Offline,
			}),
			Entry("offline when proxy is connected but all inbounds are not ready", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
										Ready: false,
									},
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Offline,
				errReasons: []string{
					"inbound[port=0,svc=] is not ready",
				},
			}),
			Entry("online when proxy is disconnected and is a gateway", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Gateway: &mesh_proto.Dataplane_Networking_Gateway{},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime:    proto.MustTimestampProto(core.Now()),
								DisconnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: Offline,
			}),
			Entry("partially degraded when proxy is connected and one inbound is ready but other is not", testCase{
				overview: &mesh_proto.DataplaneOverview{
					Dataplane: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
										Ready: true,
									},
								},
								{
									Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
										Ready: false,
									},
								},
							},
						},
					},
					DataplaneInsight: &mesh_proto.DataplaneInsight{
						Subscriptions: []*mesh_proto.DiscoverySubscription{
							{
								ConnectTime: proto.MustTimestampProto(core.Now()),
							},
						},
					},
				},
				status: PartiallyDegraded,
				errReasons: []string{
					"inbound[port=0,svc=] is not ready",
				},
			}),
		)
	})
})
