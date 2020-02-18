package topology_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/topology"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
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
			mesh := &mesh_core.MeshResource{ // mesh that is relevant to this test case
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "demo",
				},
			}
			otherMesh := &mesh_core.MeshResource{ // mesh that is irrelevant to this test case
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "default",
				},
			}
			backend := &mesh_core.DataplaneResource{ // dataplane that is a source of traffic
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "backend",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "backend", "region": "eu"},
								Port:        8080,
								ServicePort: 18080,
							},
							{
								Tags:        map[string]string{"service": "frontend", "region": "eu"},
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
			routeRedis := &mesh_core.TrafficRouteResource{ // traffic route for `redis` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "route-to-redis",
				},
				Spec: mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "frontend"}},
						{Match: mesh_proto.TagSelector{"service": "backend"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "redis"}},
					},
					Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
						{Weight: 10, Destination: mesh_proto.TagSelector{"service": "redis", "version": "v1"}},
						{Weight: 90, Destination: mesh_proto.TagSelector{"service": "redis", "version": "v2"}},
					},
				},
			}
			routeElastic := &mesh_core.TrafficRouteResource{ // traffic route for `elastic` service
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "route-to-elastic",
				},
				Spec: mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "elastic"}},
					},
					Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
						{Weight: 30, Destination: mesh_proto.TagSelector{"service": "elastic", "region": "us"}},
						{Weight: 70, Destination: mesh_proto.TagSelector{"service": "elastic", "region": "eu"}},
					},
				},
			}
			routeBlackhole := &mesh_core.TrafficRouteResource{ // traffic route that must be ignored (due to `mesh: default`)
				Meta: &test_model.ResourceMeta{
					Mesh: "default", // other mesh
					Name: "route-to-blackhole",
				},
				Spec: mesh_proto.TrafficRoute{
					Sources: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Destinations: []*mesh_proto.Selector{
						{Match: mesh_proto.TagSelector{"service": "*"}},
					},
					Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
						{Weight: 100, Destination: mesh_proto.TagSelector{"service": "blackhole"}},
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
			Expect(routes).To(HaveKey("redis"))
			Expect(routes["redis"].Meta.GetName()).To(Equal(routeRedis.Meta.GetName()))
			Expect(routes["redis"].Spec).To(Equal(routeRedis.Spec))
			// and
			Expect(routes).To(HaveKey("elastic"))
			Expect(routes["elastic"].Meta.GetName()).To(Equal(routeElastic.Meta.GetName()))
			Expect(routes["elastic"].Spec).To(Equal(routeElastic.Spec))
		})
	})

	Describe("BuildRouteMap()", func() {
		sameMeta := func(meta1, meta2 core_model.ResourceMeta) bool {
			return meta1.GetMesh() == meta2.GetMesh() &&
				meta1.GetName() == meta2.GetName() &&
				meta1.GetVersion() == meta2.GetVersion()
		}
		type testCase struct {
			dataplane *mesh_core.DataplaneResource
			routes    []*mesh_core.TrafficRouteResource
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
				Expect(routes).Should(Equal(expectedRoutes))
			},
			Entry("Dataplane without outbound interfaces and no routes", testCase{
				dataplane: &mesh_core.DataplaneResource{},
				routes:    nil,
				expected:  nil,
			}),
			Entry("Dataplane without outbound interfaces", testCase{
				dataplane: &mesh_core.DataplaneResource{},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "blackhole"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{},
			}),
			Entry("if an outbound interface has no matching TrafficRoute, an implicit default route should be used", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
								{Service: "elastic"},
							},
						},
					},
				},
				routes: nil,
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "(implicit default route)",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "redis"},
							}},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.TagSelector{"service": "redis"},
							}},
						},
					},
					"elastic": &mesh_core.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "(implicit default route)",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "elastic"},
							}},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.TagSelector{"service": "elastic"},
							}},
						},
					},
				},
			}),
			Entry("implicit default route should only be used if there is matching TrafficRoute for a given outbound interface", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
								{Service: "elastic"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-elastic-v1",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "elastic"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      200,
									Destination: mesh_proto.TagSelector{"service": "elastic", "version": "v1"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &PseudoMeta{
							Name: "(implicit default route)",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "*"},
							}},
							Destinations: []*mesh_proto.Selector{{
								Match: mesh_proto.TagSelector{"service": "redis"},
							}},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{{
								Weight:      100,
								Destination: mesh_proto.TagSelector{"service": "redis"},
							}},
						},
					},
					"elastic": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-elastic-v1",
						},
					},
				},
			}),
			Entry("TrafficRoutes should be picked by latest creation time given two equally specific routes", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "everything-to-hollygrail",
							CreationTime: time.Unix(1, 1),
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "hollygrail"},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "everything-to-blackhole",
							CreationTime: time.Unix(0, 0),
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "blackhole"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "everything-to-hollygrail",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `source` selector by 2 tags should win over a TrafficRoute with a `source` selector by 1 tag", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v1"},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend", "region": "eu"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v2"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `source` selector by an exact value should win over a TrafficRoute with a `source` selector by a wildcard value", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v1"},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v2"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("TrafficRoute with a `destination` selector by an exact value should win over a TrafficRoute with a `destination` selector by a wildcard value", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "less-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v1"},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "redis"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v2"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "more-specific",
						},
					},
				},
			}),
			Entry("in case if TrafficRoutes have equal aggregate ranks, most specific one should be selected based on last creation time", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{Tags: map[string]string{"service": "backend", "region": "eu"}},
							},
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis"},
							},
						},
					},
				},
				routes: []*mesh_core.TrafficRouteResource{
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-2",
							CreationTime: time.Unix(1, 1),
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "redis"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v2"},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Name:         "equally-specific-1",
							CreationTime: time.Unix(0, 0),
						},
						Spec: mesh_proto.TrafficRoute{
							Sources: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "backend"}},
							},
							Destinations: []*mesh_proto.Selector{
								{Match: mesh_proto.TagSelector{"service": "*"}},
							},
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "redis", "version": "v1"},
								},
							},
						},
					},
				},
				expected: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Meta: &test_model.ResourceMeta{
							Name: "equally-specific-2",
						},
					},
				},
			}),
		)
	})

	Describe("BuildDestinationMap()", func() {
		type testCase struct {
			dataplane *mesh_core.DataplaneResource
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
				dataplane: &mesh_core.DataplaneResource{},
				routes:    nil,
				expected:  core_xds.DestinationMap{},
			}),
			Entry("Dataplane with outbound interfaces but no TrafficRoutes", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
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
						{"service": "redis"},
					},
					"elastic": []mesh_proto.TagSelector{
						{"service": "elastic"},
					},
				},
			}),
			Entry("Dataplane with outbound interfaces and TrafficRoutes", testCase{
				dataplane: &mesh_core.DataplaneResource{
					Spec: mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
								{Service: "redis", Port: 10001},
								{Service: "elastic", Port: 10002},
							},
						},
					},
				},
				routes: core_xds.RouteMap{
					"redis": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      10,
									Destination: mesh_proto.TagSelector{"service": "redis", "role": "master"},
								},
								{
									Weight:      90,
									Destination: mesh_proto.TagSelector{"service": "redis", "role": "replica"},
								},
							},
						},
					},
					"elastic": &mesh_core.TrafficRouteResource{
						Spec: mesh_proto.TrafficRoute{
							Conf: []*mesh_proto.TrafficRoute_WeightedDestination{
								{
									Weight:      100,
									Destination: mesh_proto.TagSelector{"service": "google"},
								},
							},
						},
					},
				},
				expected: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"service": "redis", "role": "master"},
						{"service": "redis", "role": "replica"},
					},
					"google": []mesh_proto.TagSelector{
						{"service": "google"},
					},
				},
			}),
		)
	})
})
