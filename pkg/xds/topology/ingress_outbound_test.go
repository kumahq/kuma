package topology_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("IngressTrafficRoute", func() {
	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil, false)
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})
	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			mesh             *core_mesh.MeshResource
			dataplanes       []*core_mesh.DataplaneResource
			externalServices []*core_mesh.ExternalServiceResource
			meshServices     []*v1alpha1.MeshServiceResource
			zoneEgress       []*core_mesh.ZoneEgressResource
			expected         core_xds.EndpointMap
		}
		DescribeTable("should generate ingress outbounds matching given selectors",
			func(given testCase) {
				// when
				meshServicesByName := make(map[model.ResourceIdentifier]*v1alpha1.MeshServiceResource, len(given.meshServices))
				for _, ms := range given.meshServices {
					meshServicesByName[model.NewResourceIdentifier(ms)] = ms
				}
				endpoints := topology.BuildIngressEndpointMap(
					context.Background(),
					given.mesh,
					"east",
					meshServicesByName,
					nil,
					nil,
					given.dataplanes,
					given.externalServices,
					nil,
					given.zoneEgress,
					dataSourceLoader,
				)

				// then
				Expect(endpoints).To(Equal(given.expected))
			},

			Entry("external service for specific zone through local egress", testCase{
				mesh: samples.MeshMTLSBuilder().WithEgressRoutingEnabled().Build(),
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
					"example": []core_xds.Endpoint{
						{
							Target:          "192.168.0.1",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "example"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{},
						},
						{
							Target: "192.168.0.2",
							Port:   10002,
							Tags: map[string]string{
								mesh_proto.ServiceTag: "example",
							},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{},
						},
					},
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
							Locality:        &core_xds.Locality{Zone: "zone-2"},
						},
						{
							Target:          "192.168.0.2",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin", mesh_proto.ZoneTag: "zone-2"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
							Locality:        &core_xds.Locality{Zone: "zone-2"},
						},
					},
				},
			}),
			Entry("external service not filled when zone egress not available", testCase{
				mesh: samples.MeshDefault(),
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
				mesh: samples.MeshDefault(),
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
				mesh: samples.MeshDefault(),
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
										Port:        80,
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
						AddIntPort(8080, 80, "http").
						AddIntPort(8081, 8001, "http").
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
					"kong_kong-system_svc_80": []core_xds.Endpoint{
						{
							Target:         "192.168.0.2",
							UnixDomainPath: "",
							Port:           80,
							Tags: map[string]string{
								"kuma.io/service": "kong_kong-system_svc_80",
								"app":             "kong",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
					"default_kong.kong-system___msvc_8080": []core_xds.Endpoint{
						{
							Target:         "192.168.0.2",
							UnixDomainPath: "",
							Port:           80,
							Tags: map[string]string{
								"kuma.io/service": "kong_kong-system_svc_80",
								"app":             "kong",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
					"redis_svc_6379": []core_xds.Endpoint{
						{
							Target:         "192.168.0.1",
							UnixDomainPath: "",
							Port:           6379,
							Tags: map[string]string{
								"kuma.io/service": "redis_svc_6379",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
					"default_redis___msvc_6379": []core_xds.Endpoint{
						{
							Target:         "192.168.0.1",
							UnixDomainPath: "",
							Port:           6379,
							Tags: map[string]string{
								"kuma.io/service": "redis_svc_6379",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
					"default_redis-0___msvc_6379": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"},
							Weight: 1,
						},
					},
					"kong_kong-system_svc_8001": []core_xds.Endpoint{
						{
							Target:         "192.168.0.2",
							UnixDomainPath: "",
							Port:           8001,
							Tags: map[string]string{
								"kuma.io/service": "kong_kong-system_svc_8001",
								"app":             "kong",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
					"default_kong.kong-system___msvc_8081": []core_xds.Endpoint{
						{
							Target:         "192.168.0.2",
							UnixDomainPath: "",
							Port:           8001,
							Tags: map[string]string{
								"kuma.io/service": "kong_kong-system_svc_8001",
								"app":             "kong",
							},
							Weight:          1,
							Locality:        nil,
							ExternalService: nil,
						},
					},
				},
			}),
		)
	})
})
