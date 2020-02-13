package topology_test

import (
	"context"

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

	Describe("GetOutboundTargets()", func() {

		It("should pick proper dataplanes for each outbound destination", func() {
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
			redisV1 := &mesh_core.DataplaneResource{ // dataplane that must become a target
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "redis-v1",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.2",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "redis", "version": "v1"},
								Port:        6379,
								ServicePort: 16379,
							},
						},
					},
				},
			}
			redisV2 := &mesh_core.DataplaneResource{ // dataplane that must be ingored (due to `mesh: default`)
				Meta: &test_model.ResourceMeta{
					Mesh: "default", // other mesh
					Name: "redis-v2",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.3",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "redis", "version": "v2"},
								Port:        6379,
								ServicePort: 26379,
							},
						},
					},
				},
			}
			redisV3 := &mesh_core.DataplaneResource{ // dataplane that must be ingored (due to `version: v3`)
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "redis-v3",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.4",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "redis", "version": "v3"},
								Port:        6379,
								ServicePort: 36379,
							},
						},
					},
				},
			}
			elasticEU := &mesh_core.DataplaneResource{ // dataplane that must be ingored (due to `region: eu`)
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "elastic-eu",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.5",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "elastic", "region": "eu"},
								Port:        9200,
								ServicePort: 49200,
							},
						},
					},
				},
			}
			elasticUS := &mesh_core.DataplaneResource{ // dataplane that must become a target
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "elastic-us",
				},
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.6",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{"service": "elastic", "region": "us"},
								Port:        9200,
								ServicePort: 59200,
							},
						},
					},
				},
			}
			destinations := core_xds.DestinationMap{
				"redis": []mesh_proto.TagSelector{
					{"service": "redis", "version": "v1"},
					{"service": "redis", "version": "v2"},
				},
				"elastic": []mesh_proto.TagSelector{
					{"service": "elastic", "region": "us"},
					{"service": "elastic", "region": "au"},
				},
			}
			for _, resource := range []core_model.Resource{mesh, backend, redisV1, redisV3, elasticEU, elasticUS, otherMesh, redisV2} {
				// when
				err := rm.Create(ctx, resource, core_store.CreateBy(core_model.MetaToResourceKey(resource.GetMeta())))
				// then
				Expect(err).ToNot(HaveOccurred())
			}

			// when
			targets, err := GetOutboundTargets(ctx, backend, destinations, rm)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(targets).To(HaveLen(2))
			// and
			Expect(targets).To(HaveKeyWithValue("redis", []core_xds.Endpoint{
				{Target: "192.168.0.2", Port: 6379, Tags: map[string]string{"service": "redis", "version": "v1"}},
			}))
			Expect(targets).To(HaveKeyWithValue("elastic", []core_xds.Endpoint{
				{Target: "192.168.0.6", Port: 9200, Tags: map[string]string{"service": "elastic", "region": "us"}},
			}))
		})
	})

	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			destinations core_xds.DestinationMap
			dataplanes   []*mesh_core.DataplaneResource
			expected     core_xds.EndpointMap
		}
		DescribeTable("should include only those dataplanes that match given selectors",
			func(given testCase) {
				// when
				endpoints := BuildEndpointMap(given.destinations, given.dataplanes)
				// then
				Expect(endpoints).To(Equal(given.expected))
			},
			Entry("no destinations", testCase{
				destinations: core_xds.DestinationMap{},
				dataplanes:   []*mesh_core.DataplaneResource{},
				expected:     nil,
			}),
			Entry("no dataplanes", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{{"service": "redis"}},
				},
				dataplanes: []*mesh_core.DataplaneResource{},
				expected:   core_xds.EndpointMap{},
			}),
			Entry("no dataplanes for that service", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{{"service": "redis"}},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "elastic"},
										Port:        9200,
										ServicePort: 19200,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{},
			}),
			Entry("no dataplanes with matching tags", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{{"service": "redis", "version": "v1"}},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "version": "v2"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{},
			}),
			Entry("dataplane with invalid inbound interface", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{{"service": "redis"}},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{Interface: "invalid value", Tags: map[string]string{"service": "redis", "version": "v2"}},
								},
							},
						},
					},
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.3",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "version": "v3"},
										Port:        6379,
										ServicePort: 36379,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{Target: "192.168.0.1", Port: 6379, Tags: map[string]string{"service": "redis", "version": "v1"}},
						{Target: "192.168.0.3", Port: 6379, Tags: map[string]string{"service": "redis", "version": "v3"}},
					},
				},
			}),
			Entry("destination with multiple selectors", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"service": "redis", "region": "eu"},
						{"service": "redis", "region": "us"},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "region": "us"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{Target: "192.168.0.1", Port: 6379, Tags: map[string]string{"service": "redis", "region": "us"}},
					},
				},
			}),
			Entry("destination with multiple matching selectors", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"service": "redis"},
						{"service": "redis", "region": "us"},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "region": "us"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{Target: "192.168.0.1", Port: 6379, Tags: map[string]string{"service": "redis", "region": "us"}},
					},
				},
			}),
			Entry("multiple destinations", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"service": "redis"},
						{"service": "redis", "version": "v1"},
					},
					"elastic": []mesh_proto.TagSelector{
						{"service": "elastic"},
						{"service": "elastic", "region": "eu"},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "elastic", "region": "us"},
										Port:        9200,
										ServicePort: 19200,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{Target: "192.168.0.1", Port: 6379, Tags: map[string]string{"service": "redis", "version": "v1"}},
					},
					"elastic": []core_xds.Endpoint{
						{Target: "192.168.0.2", Port: 9200, Tags: map[string]string{"service": "elastic", "region": "us"}},
					},
				},
			}),
			Entry("multiple destinations implemented by a single dataplane", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{"service": "redis"},
						{"service": "redis", "version": "v1"},
					},
					"elastic": []mesh_proto.TagSelector{
						{"service": "elastic"},
						{"service": "elastic", "region": "eu"},
					},
				},
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"service": "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
									{
										Tags:        map[string]string{"service": "elastic", "region": "us"},
										Address:     "192.168.0.2",
										Port:        9200,
										ServicePort: 19200,
									},
								},
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{Target: "192.168.0.1", Port: 6379, Tags: map[string]string{"service": "redis", "version": "v1"}},
					},
					"elastic": []core_xds.Endpoint{
						{Target: "192.168.0.2", Port: 9200, Tags: map[string]string{"service": "elastic", "region": "us"}},
					},
				},
			}),
		)
	})
})
