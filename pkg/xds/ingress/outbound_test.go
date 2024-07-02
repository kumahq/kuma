package ingress_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	. "github.com/kumahq/kuma/pkg/xds/ingress"
)

var _ = Describe("IngressTrafficRoute", func() {
	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			destinations     core_xds.DestinationMap
			dataplanes       []*core_mesh.DataplaneResource
			externalServices []*core_mesh.ExternalServiceResource
			meshServices     []*v1alpha1.MeshServiceResource
			zoneEgress       []*core_mesh.ZoneEgressResource
			expected         core_xds.EndpointMap
		}
		DescribeTable("should generate ingress outbounds matching given selectors",
			func(given testCase) {
				// when
				endpoints := BuildEndpointMap(given.destinations, given.dataplanes, given.externalServices, given.zoneEgress, nil, given.meshServices)
				// then
				Expect(endpoints).To(Equal(given.expected))
			},

			Entry("external service for specific zone through local egress", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{mesh_proto.ServiceTag: "redis"},
					},
					"httpbin": []mesh_proto.TagSelector{
						{mesh_proto.ServiceTag: "httpbin"},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
				},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin", mesh_proto.ZoneTag: "zone-2"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "example.com:443",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "example"},
						},
					},
				},
				zoneEgress: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "ze-1",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "192.168.0.1",
								Port:    10002,
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "ze-2",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "192.168.0.2",
								Port:    10002,
							},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1, // local weight is bumped to 2 to factor two instances of Ingresses
						},
					},
					"httpbin": []core_xds.Endpoint{
						{
							Target:          "192.168.0.1",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin", mesh_proto.ZoneTag: "zone-2"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
						{
							Target:          "192.168.0.2",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin", mesh_proto.ZoneTag: "zone-2"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service not filled when zone egress not available", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{mesh_proto.ServiceTag: "redis"},
					},
					"httpbin": []mesh_proto.TagSelector{
						{mesh_proto.ServiceTag: "httpbin"},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
				},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin", mesh_proto.ZoneTag: "zone-2"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "example.com:443",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "example"},
						},
					},
				},
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1, // local weight is bumped to 2 to factor two instances of Ingresses
						},
					},
				},
			}),
			Entry("data plane proxy with advertised address", testCase{
				destinations: core_xds.DestinationMap{
					"redis": []mesh_proto.TagSelector{
						{mesh_proto.ServiceTag: "redis"},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: "default"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								AdvertisedAddress: "192.168.0.2",
								Address:           "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis"},
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
						{
							Target: "192.168.0.2",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis"},
							Weight: 1,
						},
					},
				},
			}),
			Entry("mesh services", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: model.DefaultMesh, Name: "redis-0"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: model.DefaultMesh, Name: "kong-gateway-0"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_80", "app": "kong"},
										Port:        8080,
										ServicePort: 18080,
									},
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_8001", "app": "kong"},
										Port:        8001,
										ServicePort: 18001,
									},
								},
							},
						},
					},
				},
				meshServices: []*v1alpha1.MeshServiceResource{
					builders.MeshService().
						WithName("kong.kong-system").
						WithDataplaneTagsSelectorKV("app", "kong").
						AddIntPort(80, 8080, "http").
						AddIntPort(81, 8081, "http").
						Build(),
					builders.MeshService().
						WithName("redis").
						WithDataplaneTagsSelectorKV(mesh_proto.ServiceTag, "redis_svc_6379").
						AddIntPort(6379, 6379, "tcp").
						Build(),
					builders.MeshService().
						WithName("redis-0").
						WithDataplaneRefNameSelector("redis-0").
						AddIntPort(6379, 6379, "tcp").
						Build(),
				},
				expected: core_xds.EndpointMap{
					"redis-0_svc_6379": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"},
							Weight: 1,
						},
					},
				},
			}),
		)
	})
})
