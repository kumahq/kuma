package topology_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	. "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("TrafficRoute", func() {

	var ctx context.Context
	var rm core_manager.ResourceManager

	BeforeEach(func() {
		ctx = context.Background()
		rm = core_manager.NewResourceManager(memory_resources.NewStore())
	})

	Describe("GetRoutes()", func() {

		It("should pick the best matching Route for each outbound interface", func() {
			// given
			mesh := &core_mesh.MeshResource{ // mesh that is relevant to this test case
				Meta: &test_model.ResourceMeta{
					Name: "demo",
				},
				Spec: &mesh_proto.Mesh{},
			}
			otherMesh := &core_mesh.MeshResource{ // mesh that is irrelevant to this test case
				Meta: &test_model.ResourceMeta{
					Name: "default",
				},
				Spec: &mesh_proto.Mesh{},
			}
			backend := &core_mesh.DataplaneResource{ // dataplane that is a source of traffic
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "backend",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"kuma.io/service": "backend", "region": "eu"},
								Port:        8080,
								ServicePort: 18080,
							},
							{
								Tags:        map[string]string{"kuma.io/service": "frontend", "region": "eu"},
								Port:        7070,
								ServicePort: 17070,
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{Service: "redis", Port: 10001},
							{Service: "elastic", Port: 10002},
						},
					},
				},
			}
			routeRedis := &core_mesh.TrafficRouteResource{ // traffic route for `redis` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "route-to-redis",
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "frontend"}},
						{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
					},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Split: []*mesh_proto.TrafficRoute_Split{
							{
								Weight:      util_proto.UInt32(10),
								Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v1"},
							},
							{
								Weight:      util_proto.UInt32(90),
								Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v2"},
							},
						},
					},
				},
			}
			routeElastic := &core_mesh.TrafficRouteResource{ // traffic route for `elastic` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "route-to-elastic",
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"}},
					},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Split: []*mesh_proto.TrafficRoute_Split{
							{
								Weight:      util_proto.UInt32(30),
								Destination: mesh_proto.TagSelector{"kuma.io/service": "elastic", "region": "us"},
							},
							{
								Weight:      util_proto.UInt32(70),
								Destination: mesh_proto.TagSelector{"kuma.io/service": "elastic", "region": "eu"},
							},
						},
					},
				},
			}
			routeBlackhole := &core_mesh.TrafficRouteResource{ // traffic route that must be ignored (due to `mesh: default`)
				Meta: &test_model.ResourceMeta{
					Mesh: "default", // other mesh
					Name: "route-to-blackhole",
				},
				Spec: &mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
					},
					Conf: &mesh_proto.TrafficRoute_Conf{
						Split: []*mesh_proto.TrafficRoute_Split{
							{
								Weight:      util_proto.UInt32(100),
								Destination: mesh_proto.TagSelector{"kuma.io/service": "blackhole"},
							},
						},
					},
				},
			}
			for _, resource := range []core_model.Resource{mesh, backend, routeRedis, routeElastic, otherMesh, routeBlackhole} {
				// when
				err := rm.Create(ctx, resource, core_store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				// then
				Expect(err).ToNot(HaveOccurred())
			}

			// when
			routes, err := GetRoutes(ctx, backend, rm)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(routes).To(HaveLen(2))
			// and
			out1 := mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: 10001}
			Expect(routes).To(HaveKey(out1))
			Expect(routes[out1].Meta.GetName()).To(Equal(routeRedis.Meta.GetName()))
			Expect(routes[out1].Spec).To(MatchProto(routeRedis.Spec))
			// and
			out2 := mesh_proto.OutboundInterface{DataplaneIP: "127.0.0.1", DataplanePort: 10002}
			Expect(routes).To(HaveKey(out2))
			Expect(routes[out2].Meta.GetName()).To(Equal(routeElastic.Meta.GetName()))
			Expect(routes[out2].Spec).To(MatchProto(routeElastic.Spec))
		})
	})

	Describe("BuildRouteMap()", func() {
		sameMeta := func(meta1, meta2 core_model.ResourceMeta) bool {
			return meta1.GetMesh() == meta2.GetMesh() &&
				meta1.GetName() == meta2.GetName() &&
				meta1.GetVersion() == meta2.GetVersion()
		}
		type testCase struct {
			dataplane *core_mesh.DataplaneResource
			routes    []*core_mesh.TrafficRouteResource
			expected  core_xds.RouteMap
		}
		DescribeTable("should correctly pick a single the most specific route for each outbound interface",
			func(given testCase) {
				// setup
				expectedRoutes := core_xds.RouteMap{}
				for service, expectedRoute := range given.expected {
					for _, route := range given.routes {
						if sameMeta(expectedRoute.GetMeta(), route.GetMeta()) {
							expectedRoutes[service] = route
							break
						}
					}
					if _, ok := expectedRoutes[service]; !ok {
						expectedRoutes[service] = expectedRoute
					}
				}
				// when
				routes := BuildRouteMap(given.dataplane, given.routes)
				// expect
				Expect(routes).To(HaveLen(len(expectedRoutes)))
				for outbound, trafficRouteRes := range expectedRoutes {
					Expect(routes[outbound].Meta).To(Equal(trafficRouteRes.Meta))
					Expect(routes[outbound].Spec).To(MatchProto(trafficRouteRes.Spec))
				}
			},
			Entry("Dataplane without outbound interfaces and no routes", testCase{
				dataplane: core_mesh.NewDataplaneResource(),
				routes:    nil,
				expected:  nil,
			}),
			Entry("Dataplane without outbound interfaces", testCase{
				dataplane: core_mesh.NewDataplaneResource(),
				routes: []*core_mesh.TrafficRouteResource{
					{
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "blackhole"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{},
			}),
			Entry("TrafficRoutes should be picked by latest creation time given two equally specific routes", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "everything-to-hollygrail",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "hollygrail"},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "everything-to-blackhole",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "blackhole"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-hollygrail",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `source` selector by 2 tags should win over a TrafficRoute with a `source` selector by 1 tag", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v1"},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "backend", "region": "eu"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v2"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `source` selector by an exact value should win over a TrafficRoute with a `source` selector by a wildcard value", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v1"},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v2"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `destination` selector by an exact value should win over a TrafficRoute with a `destination` selector by a wildcard value", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v1"},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v2"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("in case if TrafficRoutes have equal aggregate ranks, most specific one should be selected based on last creation time", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-2",
							CreationTime: time.Unix(1, 1),
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "redis"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v2"},
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-1",
							CreationTime: time.Unix(0, 0),
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "version": "v1"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "equally-specific-2",
						},
					},
				},
			}),
		)
		DescribeTable("should pick the default route",
			func(given testCase) {
				// when
				routes := BuildRouteMap(given.dataplane, given.routes)
				// expect
				Expect(routes).Should(Equal(given.expected))
			},
			Entry("if an outbound interface has no matching TrafficRoute, the default route should be used", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
								{
									Port:    1235,
									Service: "elastic",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &PseudoMeta{
							Name: "route-all-default",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							},
							},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{{
									Weight:      util_proto.UInt32(100),
									Destination: mesh_proto.MatchAnyService(),
								},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "route-all-default",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis"},
									},
								},
							},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1235,
					}: &core_mesh.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "route-all-default",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "elastic"},
									},
								},
							},
						},
					},
				},
			}),
			Entry("the default route should only be used if there is matching TrafficRoute for a given outbound interface", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"kuma.io/service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{
									Port:    1234,
									Service: "redis",
								},
								{
									Port:    1235,
									Service: "elastic",
								},
							},
						},
					},
				},
				routes: []*core_mesh.TrafficRouteResource{
					{
						Meta: &PseudoMeta{
							Name: "route-all-default",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							},
							},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.MatchAnyService(),
							},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{{
									Weight:      util_proto.UInt32(100),
									Destination: mesh_proto.MatchAnyService(),
								},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-elastic-v1",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"}},
							},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "elastic", "version": "v1"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1234,
					}: &core_mesh.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "route-all-default",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis"},
									},
								},
							},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 1235,
					}: &core_mesh.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-elastic-v1",
						},
						Spec: &mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"},
							}},
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "elastic", "version": "v1"},
									},
								},
							},
						},
					},
				},
			}),
		)
	})

	Describe("BuildDestinationMap()", func() {
		type testCase struct {
			dataplane *core_mesh.DataplaneResource
			routes    core_xds.RouteMap
			expected  core_xds.DestinationMap
		}
		DescribeTable("should correctly combine outbound interfaces with configuration of TrafficRoutes",
			func(given testCase) {
				// when
				destinations := BuildDestinationMap(given.dataplane, given.routes)
				// expect
				Expect(destinations).Should(Equal(given.expected))
			},
			Entry("Dataplane without outbound interfaces", testCase{
				dataplane: core_mesh.NewDataplaneResource(),
				routes:    nil,
				expected:  core_xds.DestinationMap{},
			}),
			Entry("Dataplane with outbound interfaces but no TrafficRoutes", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis", Port: 10001},
								{Service: "elastic", Port: 10002},
							},
						},
					},
				},
				routes: nil,
				expected: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"kuma.io/service": "redis"},
					},
					"elastic": []mesh_proto.TagSelector{
						{"kuma.io/service": "elastic"},
					},
				},
			}),
			Entry("Dataplane with outbound interfaces and TrafficRoutes", testCase{
				dataplane: &core_mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis", Port: 10001},
								{Service: "elastic", Port: 10002},
							},
						},
					},
				},
				routes: core_xds.RouteMap{
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 10001,
					}: &core_mesh.TrafficRouteResource{
						Spec: &mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "role": "master"},
									},
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "redis", "role": "replica"},
									},
								},
							},
						},
					},
					mesh_proto.OutboundInterface{
						DataplaneIP:   "127.0.0.1",
						DataplanePort: 10002,
					}: &core_mesh.TrafficRouteResource{
						Spec: &mesh_proto.TrafficRoute{
							Conf: &mesh_proto.TrafficRoute_Conf{
								Split: []*mesh_proto.TrafficRoute_Split{
									{
										Weight:      util_proto.UInt32(100),
										Destination: mesh_proto.TagSelector{"kuma.io/service": "google"},
									},
								},
							},
						},
					},
				},
				expected: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"kuma.io/service": "redis", "role": "master"},
						{"kuma.io/service": "redis", "role": "replica"},
					},
					"google": []mesh_proto.TagSelector{
						{"kuma.io/service": "google"},
					},
				},
			}),
		)
	})
})
