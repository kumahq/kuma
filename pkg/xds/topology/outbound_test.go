package topology_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("TrafficRoute", func() {
	const defaultMeshName = "default"
	defaultMeshWithMTLS := &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{
			Name: defaultMeshName,
		},
		Spec: &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
			},
		},
	}
	defaultMeshWithMTLSAndZoneEgress := &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{
			Name: defaultMeshName,
		},
		Spec: &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
			},
			Routing: &mesh_proto.Routing{
				ZoneEgress: true,
			},
		},
	}
	defaultMeshWithMTLSAndZoneEgressDisabled := &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{
			Name: defaultMeshName,
		},
		Spec: &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
			},
			Routing: &mesh_proto.Routing{
				ZoneEgress: false,
			},
		},
	}
	defaultMeshWithoutMTLS := &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{
			Name: defaultMeshName,
		},
		Spec: &mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "",
			},
		},
	}
	defaultMeshWithLocality := &core_mesh.MeshResource{
		Meta: &test_model.ResourceMeta{
			Name: defaultMeshName,
		},
		Spec: &mesh_proto.Mesh{
			Routing: &mesh_proto.Routing{
				LocalityAwareLoadBalancing: true,
			},
		},
	}
	const nonDefaultMesh = "non-default"

	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil, false)
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})
	Describe("GetOutboundTargets()", func() {
		It("should pick proper dataplanes for each outbound destination", func() {
			// given
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
								Tags:        map[string]string{mesh_proto.ServiceTag: "backend", mesh_proto.ZoneTag: "eu"},
								Port:        8080,
								ServicePort: 18080,
							},
							{
								Tags:        map[string]string{mesh_proto.ServiceTag: "frontend", mesh_proto.ZoneTag: "eu"},
								Port:        7070,
								ServicePort: 17070,
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Tags: map[string]string{
									mesh_proto.ServiceTag: "redis",
								},
								Port: 10001,
							},
							{
								Tags: map[string]string{
									mesh_proto.ServiceTag: "elastic",
								},
								Port: 10002,
							},
						},
					},
				},
			}
			redisV1 := &core_mesh.DataplaneResource{ // dataplane that must become a target
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "redis-v1",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.2",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
								Port:        6379,
								ServicePort: 16379,
							},
						},
					},
				},
			}
			redisV3 := &core_mesh.DataplaneResource{ // dataplane that must be ingored (due to `version: v3`)
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "redis-v3",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.4",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
								Port:        6379,
								ServicePort: 36379,
							},
						},
					},
				},
			}
			elasticEU := &core_mesh.DataplaneResource{ // dataplane that must be ingored (due to `kuma.io/zone: eu`)
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "elastic-eu",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.5",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{mesh_proto.ServiceTag: "elastic", mesh_proto.ZoneTag: "eu"},
								Port:        9200,
								ServicePort: 49200,
							},
						},
					},
				},
			}
			elasticUS := &core_mesh.DataplaneResource{ // dataplane that must become a target
				Meta: &test_model.ResourceMeta{
					Mesh: "demo",
					Name: "elastic-us",
				},
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.6",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Tags:        map[string]string{mesh_proto.ServiceTag: "elastic", mesh_proto.ZoneTag: "us"},
								Port:        9200,
								ServicePort: 59200,
							},
						},
					},
				},
			}
			dataplanes := &core_mesh.DataplaneResourceList{
				Items: []*core_mesh.DataplaneResource{backend, redisV1, redisV3, elasticEU, elasticUS},
			}

			externalServices := &core_mesh.ExternalServiceResourceList{}

			// when
			targets := BuildEdsEndpointMap(
				defaultMeshWithMTLS, "zone-1", nil, nil, nil, dataplanes.Items, nil, nil, externalServices.Items,
			)

			Expect(targets).To(HaveLen(4))
			// and
			Expect(targets).To(HaveKeyWithValue("redis", []core_xds.Endpoint{
				{
					Target: "192.168.0.2",
					Port:   6379,
					Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
					Weight: 1,
				},
				{
					Target: "192.168.0.4",
					Port:   6379,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "redis",
						"version":             "v3",
					},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("elastic", []core_xds.Endpoint{
				{
					Target: "192.168.0.5",
					Port:   9200,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "elastic",
						mesh_proto.ZoneTag:    "eu",
					},
					Locality: &core_xds.Locality{
						Zone: "eu",
					},
					Weight: 1,
				},
				{
					Target: "192.168.0.6",
					Port:   9200,
					Tags:   map[string]string{mesh_proto.ServiceTag: "elastic", mesh_proto.ZoneTag: "us"},
					Locality: &core_xds.Locality{
						Zone: "us",
					},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("frontend", []core_xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   7070,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "frontend",
						mesh_proto.ZoneTag:    "eu",
					},
					Locality: &core_xds.Locality{
						Zone: "eu",
					},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("backend", []core_xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   8080,
					Tags: map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    "eu",
					},
					Locality: &core_xds.Locality{
						Zone: "eu",
					},
					Weight: 1,
				},
			}))
		})
	})

	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			dataplanes           []*core_mesh.DataplaneResource
			meshServices         []*meshservice_api.MeshServiceResource
			meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource
			meshMultiZoneService []*meshmzservice_api.MeshMultiZoneServiceResource
			zoneIngresses        []*core_mesh.ZoneIngressResource
			zoneEgresses         []*core_mesh.ZoneEgressResource
			externalServices     []*core_mesh.ExternalServiceResource
			mesh                 *core_mesh.MeshResource
			expected             core_xds.EndpointMap
		}
		DescribeTable("should include only those dataplanes that match given selectors",
			func(given testCase) {
				// when
				meshServiceByName := map[core_model.ResourceIdentifier]*meshservice_api.MeshServiceResource{}
				for _, ms := range given.meshServices {
					meshServiceByName[core_model.NewResourceIdentifier(ms)] = ms
				}
				endpoints := BuildEdsEndpointMap(
					given.mesh,
					"zone-1",
					meshServiceByName,
					given.meshMultiZoneService,
					given.meshExternalServices,
					given.dataplanes,
					given.zoneIngresses,
					given.zoneEgresses,
					given.externalServices,
				)
				esEndpoints := BuildExternalServicesEndpointMap(
					context.Background(), given.mesh, given.externalServices, dataSourceLoader, "zone-1",
				)
				for k, v := range esEndpoints {
					endpoints[k] = v
				}
				// then
				Expect(endpoints).To(Equal(given.expected))
			},
			Entry("no dataplanes", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				mesh:       defaultMeshWithMTLS,
				expected:   core_xds.EndpointMap{},
			}),
			Entry("ingress in the list of dataplanes", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.2",
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances: 2,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
								},
								{
									Instances: 3,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
								},
							},
						},
					},
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.3", // another instance of the same ingress will be ignored
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							// when AvailableServices are not computed for the instance of the ingress behind the same
							// load balancer (advertised address + port), available services from an instance that has them
							// is preferred.
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{},
						},
					},
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.4",
								Port:              10001,
								AdvertisedAddress: "192.168.0.101",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances: 2,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
								},
								{
									Instances: 3,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
								},
							},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
							Locality: &core_xds.Locality{
								Zone: "eu",
							},
							Weight: 2,
						},
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
							Weight: 3,
						},
						{
							Target: "192.168.0.101",
							Port:   12345,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
							Locality: &core_xds.Locality{
								Zone: "eu",
							},
							Weight: 2,
						},
						{
							Target: "192.168.0.101",
							Port:   12345,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
							Weight: 3,
						},
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 2, // local weight is bumped to 2 to factor two instances of Ingresses
						},
					},
				},
			}),
			Entry("ingresses in the list of dataplanes from different meshes", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.2",
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances: 2,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
								},
								{
									Instances: 3,
									Mesh:      nonDefaultMesh,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
								},
							},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
							Locality: &core_xds.Locality{
								Zone: "eu",
							},
							Weight: 2,
						},
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1,
						},
					},
				},
			}),
			Entry("ingress is not included if mtls is off", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address: "10.20.1.2",
								Port:    10001,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances: 2,
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", mesh_proto.ZoneTag: "eu"},
								},
								{
									Instances: 3,
									Mesh:      nonDefaultMesh,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
								},
							},
						},
					},
				},
				mesh: defaultMeshWithoutMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1,
						},
					},
				},
			}),
			Entry("external service no TLS", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls:     nil,
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service with TLS disabled", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									Enabled: false,
								},
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service with TLS enabled", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									Enabled: true,
								},
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ExternalServiceTag: ""},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: true},
						},
					},
				},
			}),
			Entry("external service with TLS enabled and Locality", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									Enabled: true,
								},
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "us"},
						},
					},
				},
				mesh: defaultMeshWithLocality,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "us", mesh_proto.ExternalServiceTag: ""},
							Weight:          1,
							Locality:        &core_xds.Locality{Zone: "us", Priority: 1},
							ExternalService: &core_xds.ExternalService{TLSEnabled: true},
						},
					},
				},
			}),
			Entry("external services with Zones and Locality", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				externalServices: []*core_mesh.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "zone1.httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "zone-1"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "zone2.httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "zone-2"},
						},
					},
				},
				mesh: defaultMeshWithLocality,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "zone1.httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "zone-1"},
							Weight:          1,
							Locality:        &core_xds.Locality{Zone: "zone-1", Priority: 0},
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
						{
							Target:          "zone2.httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "redis", mesh_proto.ZoneTag: "zone-2"},
							Weight:          1,
							Locality:        &core_xds.Locality{Zone: "zone-2", Priority: 1},
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("unhealthy dataplane", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Name: "dp-1", Mesh: defaultMeshName},
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
					{
						Meta: &test_model.ResourceMeta{Name: "dp-2", Mesh: defaultMeshName},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
										Health:      &mesh_proto.Dataplane_Networking_Inbound_Health{Ready: false},
									},
								},
							},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
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
			Entry("external service with zoneegress address when mtls and zone egress enabled and zone egress instance available", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "example.com:443",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "example"},
						},
					},
				},
				zoneEgresses: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "egress",
							Mesh: "default",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "1.1.1.1",
								Port:    10002,
							},
						},
					},
				},
				mesh: defaultMeshWithMTLSAndZoneEgress,
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
							Target:          "1.1.1.1",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
					"example": []core_xds.Endpoint{
						{
							Target:          "1.1.1.1",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "example"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service with direct address when mtls and zone egress disabled", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "example.com:443",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "example"},
						},
					},
				},
				zoneEgresses: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "egress",
							Mesh: "default",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "1.1.1.1",
								Port:    10002,
							},
						},
					},
				},
				mesh: defaultMeshWithMTLSAndZoneEgressDisabled,
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
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
					"example": []core_xds.Endpoint{
						{
							Target:          "example.com",
							Port:            443,
							Tags:            map[string]string{mesh_proto.ServiceTag: "example"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("no external services when mtls and zone egress enabled but no zone egress instance", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin"},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "example.com:443",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "example"},
						},
					},
				},
				mesh: defaultMeshWithMTLSAndZoneEgress,
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
			Entry("external services available from other zone should have non empty external service object for dp", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.2",
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances:       2,
									Mesh:            defaultMeshName,
									Tags:            map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2"},
									ExternalService: true,
								},
							},
						},
					},
				},
				zoneEgresses: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "egress",
							Mesh: "default",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "1.1.1.1",
								Port:    10002,
							},
						},
					},
				},
				mesh: defaultMeshWithMTLSAndZoneEgress,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1, // local weight is bumped to 2 to factor two instances of Ingresses
						},
					},
					"service-in-zone2": []core_xds.Endpoint{
						{
							Target:          "1.1.1.1",
							Port:            10002,
							Tags:            map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2"},
							Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
							Locality:        &core_xds.Locality{Zone: "zone-2", Priority: 0},
							ExternalService: &core_xds.ExternalService{},
						},
					},
				},
			}),
			Entry("service in zone2 available through ingress when zoneEgress disabled but zoneEgress instances available", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
							},
							Tags: map[string]string{mesh_proto.ServiceTag: "httpbin"},
						},
					},
				},
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.2",
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{
									Instances:       2,
									Mesh:            defaultMeshName,
									Tags:            map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2"},
									ExternalService: true,
								},
							},
						},
					},
				},
				zoneEgresses: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "egress",
							Mesh: "default",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "1.1.1.1",
								Port:    10002,
							},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{mesh_proto.ServiceTag: "redis", "version": "v1"},
							Weight: 1,
						},
					},
					"service-in-zone2": []core_xds.Endpoint{
						{
							Target:   "192.168.0.100",
							Port:     12345,
							Tags:     map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2"},
							Weight:   2, // local weight is bumped to 2 to factor two instances of Ingresses
							Locality: &core_xds.Locality{Zone: "zone-2"},
						},
					},
					"httpbin": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{},
						},
					},
				},
			}),
			Entry("uses MeshService", testCase{
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: &mesh_proto.ZoneIngress{
							Zone: "zone-2",
							Networking: &mesh_proto.ZoneIngress_Networking{
								Address:           "10.20.1.2",
								Port:              10001,
								AdvertisedAddress: "192.168.0.100",
								AdvertisedPort:    12345,
							},
							AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
								{ // should ignore this because we prefer MeshService routing
									Mesh:      defaultMeshName,
									Tags:      map[string]string{mesh_proto.ServiceTag: "redis_svc_6379", "version": "v1"},
									Instances: 1,
								},
							},
						},
					},
				},
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName, Name: "redis-0"},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{mesh_proto.ServiceTag: "redis_svc_6379", "version": "v1"},
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
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
				meshServices: []*meshservice_api.MeshServiceResource{
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
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis_svc_6379": []core_xds.Endpoint{
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags: map[string]string{
								"kuma.io/service": "redis_svc_6379",
								"version":         "v1",
							},
							Weight:   1,
							Locality: nil,
						},
						{
							Target:   "192.168.0.1",
							Port:     6379,
							Tags:     map[string]string{mesh_proto.ServiceTag: "redis_svc_6379", "version": "v1"},
							Locality: nil,
							Weight:   1,
						},
					},
					"default_redis___msvc_6379": []core_xds.Endpoint{
						{
							Target:   "192.168.0.1",
							Port:     6379,
							Tags:     map[string]string{mesh_proto.ServiceTag: "redis_svc_6379", "version": "v1"},
							Locality: nil,
							Weight:   1,
						},
					},
					"default_redis-0___msvc_6379": []core_xds.Endpoint{
						{
							Target:   "192.168.0.1",
							Port:     6379,
							Tags:     map[string]string{mesh_proto.ServiceTag: "redis_svc_6379", "version": "v1"},
							Locality: nil,
							Weight:   1,
						},
					},
					"kong_kong-system_svc_80": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     80,
							Tags:     map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_80", "app": "kong"},
							Locality: nil,
							Weight:   1,
						},
					},
					"default_kong.kong-system___msvc_8080": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     80,
							Tags:     map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_80", "app": "kong"},
							Locality: nil,
							Weight:   1,
						},
					},
					"kong_kong-system_svc_8001": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     8001,
							Tags:     map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_8001", "app": "kong"},
							Locality: nil,
							Weight:   1,
						},
					},
					"default_kong.kong-system___msvc_8081": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     8001,
							Tags:     map[string]string{mesh_proto.ServiceTag: "kong_kong-system_svc_8001", "app": "kong"},
							Locality: nil,
							Weight:   1,
						},
					},
				},
			}),
			Entry("uses MeshExternalService with egress", testCase{
				meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "example-mes",
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
								Port:     10000,
								Protocol: core_mesh.ProtocolHTTP,
							},
							Endpoints: []meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    meshexternalservice_api.Port(443),
								},
							},
							Tls: &meshexternalservice_api.Tls{
								Enabled: true,
								Version: &common_tls.Version{
									Min: pointer.To(common_tls.TLSVersion12),
									Max: pointer.To(common_tls.TLSVersion13),
								},
								AllowRenegotiation: true,
								Verification: &meshexternalservice_api.Verification{
									Mode:       pointer.To(meshexternalservice_api.TLSVerificationSecured),
									ServerName: pointer.To("example.com"),
									SubjectAltNames: &[]meshexternalservice_api.SANMatch{
										{
											Type:  meshexternalservice_api.SANMatchPrefix,
											Value: "test.com",
										},
										{
											Type:  meshexternalservice_api.SANMatchExact,
											Value: "test.com",
										},
									},
									CaCert: &common_api.DataSource{
										InlineString: pointer.To("ca"),
									},
									ClientCert: &common_api.DataSource{
										InlineString: pointer.To("cert"),
									},
									ClientKey: &common_api.DataSource{
										InlineString: pointer.To("key"),
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "another-mes",
							Labels: map[string]string{
								"custom-label": "label",
							},
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
								Port:     10000,
								Protocol: core_mesh.ProtocolTCP,
							},
							Endpoints: []meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    meshexternalservice_api.Port(443),
								},
							},
							Tls: &meshexternalservice_api.Tls{
								Enabled: true,
								Verification: &meshexternalservice_api.Verification{
									Mode:       pointer.To(meshexternalservice_api.TLSVerificationSkipSAN),
									ServerName: pointer.To("example.com"),
									SubjectAltNames: &[]meshexternalservice_api.SANMatch{
										{
											Type:  meshexternalservice_api.SANMatchPrefix,
											Value: "test.com",
										},
										{
											Type:  meshexternalservice_api.SANMatchExact,
											Value: "test.com",
										},
									},
								},
							},
						},
					},
				},
				zoneEgresses: []*core_mesh.ZoneEgressResource{
					{
						Meta: &test_model.ResourceMeta{
							Name: "egress",
							Mesh: "default",
						},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{
								Address: "1.1.1.1",
								Port:    10002,
							},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"default_another-mes___extsvc_10000": []core_xds.Endpoint{
						{
							Target: "1.1.1.1",
							Port:   10002,
							Tags: map[string]string{
								"custom-label": "label",
							},
							Locality: nil,
							Weight:   1,
							ExternalService: &core_xds.ExternalService{
								Protocol:   core_mesh.ProtocolTCP,
								TLSEnabled: false,
								OwnerResource: &core_model.TypedResourceIdentifier{
									ResourceIdentifier: core_model.ResourceIdentifier{
										Name: "another-mes",
										Mesh: "default",
									},
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
								},
							},
						},
					},
					"default_example-mes___extsvc_10000": []core_xds.Endpoint{
						{
							Target:   "1.1.1.1",
							Port:     10002,
							Locality: nil,
							Weight:   1,
							ExternalService: &core_xds.ExternalService{
								Protocol:   core_mesh.ProtocolHTTP,
								TLSEnabled: false,
								OwnerResource: &core_model.TypedResourceIdentifier{
									ResourceIdentifier: core_model.ResourceIdentifier{
										Name: "example-mes",
										Mesh: "default",
									},
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
								},
							},
						},
					},
				},
			}),
			Entry("uses MeshExternalService without egress", testCase{
				meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "example-mes",
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
								Port:     10000,
								Protocol: core_mesh.ProtocolHTTP,
							},
							Endpoints: []meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    meshexternalservice_api.Port(443),
								},
							},
							Tls: &meshexternalservice_api.Tls{
								Enabled: true,
								Version: &common_tls.Version{
									Min: pointer.To(common_tls.TLSVersion12),
									Max: pointer.To(common_tls.TLSVersion13),
								},
								AllowRenegotiation: true,
								Verification: &meshexternalservice_api.Verification{
									Mode:       pointer.To(meshexternalservice_api.TLSVerificationSecured),
									ServerName: pointer.To("example.com"),
									SubjectAltNames: &[]meshexternalservice_api.SANMatch{
										{
											Type:  meshexternalservice_api.SANMatchPrefix,
											Value: "test.com",
										},
										{
											Type:  meshexternalservice_api.SANMatchExact,
											Value: "test.com",
										},
									},
									CaCert: &common_api.DataSource{
										InlineString: pointer.To("ca"),
									},
									ClientCert: &common_api.DataSource{
										InlineString: pointer.To("cert"),
									},
									ClientKey: &common_api.DataSource{
										InlineString: pointer.To("key"),
									},
								},
							},
						},
					},
				},
				mesh:     defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{},
			}),
			Entry("uses MeshMultiZoneService", testCase{
				zoneIngresses: []*core_mesh.ZoneIngressResource{
					builders.ZoneIngress().
						WithZone("east").
						WithAdvertisedAddress("192.168.0.100").
						WithAdvertisedPort(12345).
						Build(),
				},
				dataplanes: []*core_mesh.DataplaneResource{
					samples.DataplaneBackend(),
				},
				meshServices: []*meshservice_api.MeshServiceResource{
					samples.MeshServiceBackend(),
					samples.MeshServiceSyncedBackend(),
				},
				meshMultiZoneService: []*meshmzservice_api.MeshMultiZoneServiceResource{
					samples.MeshMultiZoneServiceBackendBuilder().
						AddMatchedMeshServiceName(core_model.NewResourceIdentifier(samples.MeshServiceBackend())).
						AddMatchedMeshServiceName(core_model.NewResourceIdentifier(samples.MeshServiceSyncedBackend())).
						Build(),
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"backend": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   80,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
							Weight: 1,
						},
					},
					"default_backend___msvc_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   80,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
							Weight: 1,
						},
					},
					"default_backend__east_msvc_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags: map[string]string{
								"kuma.io/service": "default_backend__east_msvc_80",
								"kuma.io/zone":    "east",
							},
							Weight:   1,
							Locality: &core_xds.Locality{Zone: "east", SubZone: "", Priority: 1, Weight: 0},
						},
					},
					"default_backend___mzsvc_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   80,
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
							Weight: 1,
						},
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags: map[string]string{
								"kuma.io/service": "default_backend__east_msvc_80",
								"kuma.io/zone":    "east",
							},
							Weight:   1,
							Locality: &core_xds.Locality{Zone: "east", SubZone: "", Priority: 1, Weight: 0},
						},
					},
				},
			}),
		)
		Describe("BuildEgressEndpointMap()", func() {
			type testCase struct {
				zoneIngresses        []*core_mesh.ZoneIngressResource
				externalServices     []*core_mesh.ExternalServiceResource
				meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource
				mesh                 *core_mesh.MeshResource
				expected             core_xds.EndpointMap
			}

			DescribeTable("should generate endpoints map for zone egress",
				func(given testCase) {
					// when
					endpoints := BuildEgressEndpointMap(context.Background(), given.mesh, "zone-1", given.zoneIngresses, given.externalServices, given.meshExternalServices, dataSourceLoader)
					// then
					Expect(endpoints).To(Equal(given.expected))
				},
				Entry("generate map for zone egress with ingress instances", testCase{
					zoneIngresses: []*core_mesh.ZoneIngressResource{
						{
							Spec: &mesh_proto.ZoneIngress{
								Zone: "zone-2",
								Networking: &mesh_proto.ZoneIngress_Networking{
									Address:           "10.20.1.2",
									Port:              10001,
									AdvertisedAddress: "192.168.0.100",
									AdvertisedPort:    12345,
								},
								AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
									{
										Instances:       2,
										Mesh:            defaultMeshName,
										Tags:            map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2"},
										ExternalService: true,
									},
									{
										Instances: 3,
										Mesh:      defaultMeshName,
										Tags:      map[string]string{mesh_proto.ServiceTag: "test", mesh_proto.ZoneTag: "zone-2"},
									},
								},
							},
						},
					},
					externalServices: []*core_mesh.ExternalServiceResource{
						{
							Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
							Spec: &mesh_proto.ExternalService{
								Networking: &mesh_proto.ExternalService_Networking{
									Address: "httpbin.org:80",
								},
								Tags: map[string]string{mesh_proto.ServiceTag: "httpbin"},
							},
						},
					},
					meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
						{
							Meta: &test_model.ResourceMeta{Mesh: defaultMeshName, Name: "example"},
							Spec: &meshexternalservice_api.MeshExternalService{
								Match: meshexternalservice_api.Match{
									Type:     pointer.To(meshexternalservice_api.HostnameGeneratorType),
									Port:     443,
									Protocol: core_mesh.ProtocolTCP,
								},
								Endpoints: []meshexternalservice_api.Endpoint{
									{
										Address: "192.168.1.1",
										Port:    meshexternalservice_api.Port(10000),
									},
								},
							},
						},
					},
					mesh: defaultMeshWithMTLSAndZoneEgress,
					expected: core_xds.EndpointMap{
						"service-in-zone2": []core_xds.Endpoint{
							{
								Target:   "192.168.0.100",
								Port:     12345,
								Tags:     map[string]string{mesh_proto.ServiceTag: "service-in-zone2", mesh_proto.ZoneTag: "zone-2", "mesh": "default"},
								Weight:   2, // local weight is bumped to 2 to factor two instances of Ingresses
								Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
							},
						},
						"test": []core_xds.Endpoint{
							{
								Target:   "192.168.0.100",
								Port:     12345,
								Tags:     map[string]string{mesh_proto.ServiceTag: "test", mesh_proto.ZoneTag: "zone-2", "mesh": "default"},
								Weight:   3, // local weight is bumped to 2 to factor two instances of Ingresses
								Locality: &core_xds.Locality{Zone: "zone-2", Priority: 0},
							},
						},
						"httpbin": []core_xds.Endpoint{
							{
								Target:          "httpbin.org",
								Port:            80,
								Tags:            map[string]string{mesh_proto.ServiceTag: "httpbin", "mesh": "default"},
								Weight:          1, // local weight is bumped to 2 to factor two instances of Ingresses
								ExternalService: &core_xds.ExternalService{TLSEnabled: false},
							},
						},
						"default_example___extsvc_443": []core_xds.Endpoint{
							{
								Target: "192.168.1.1",
								Port:   10000,
								Tags: map[string]string{
									"mesh": "default",
								},
								Weight: 1,
								ExternalService: &core_xds.ExternalService{
									TLSEnabled: false,
									Protocol:   core_mesh.ProtocolTCP,
									OwnerResource: &core_model.TypedResourceIdentifier{
										ResourceType: "MeshExternalService",
										ResourceIdentifier: core_model.ResourceIdentifier{
											Name: "example",
											Mesh: "default",
										},
									},
								},
							},
						},
					},
				}),
			)
		})
	})
})
