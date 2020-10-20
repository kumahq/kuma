package topology_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	. "github.com/kumahq/kuma/pkg/xds/topology"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("TrafficRoute", func() {
	const defaultMeshName = "default"
	defaultMeshWithMTLS := &mesh_core.MeshResource{
		Meta: &test_model.ResourceMeta{
			Mesh: defaultMeshName,
			Name: defaultMeshName,
		},
		Spec: mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "ca-1",
			},
		},
	}
	defaultMeshWithoutMTLS := &mesh_core.MeshResource{
		Meta: &test_model.ResourceMeta{
			Mesh: defaultMeshName,
			Name: defaultMeshName,
		},
		Spec: mesh_proto.Mesh{
			Mtls: &mesh_proto.Mesh_Mtls{
				EnabledBackend: "",
			},
		},
	}
	const nonDefaultMesh = "non-default"

	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil)
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})
	Describe("GetOutboundTargets()", func() {

		It("should pick proper dataplanes for each outbound destination", func() {
			// given
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
								Tags:        map[string]string{"kuma.io/service": "redis", "version": "v1"},
								Port:        6379,
								ServicePort: 16379,
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
								Tags:        map[string]string{"kuma.io/service": "redis", "version": "v3"},
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
								Tags:        map[string]string{"kuma.io/service": "elastic", "region": "eu"},
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
								Tags:        map[string]string{"kuma.io/service": "elastic", "region": "us"},
								Port:        9200,
								ServicePort: 59200,
							},
						},
					},
				},
			}
			dataplanes := &mesh_core.DataplaneResourceList{
				Items: []*mesh_core.DataplaneResource{backend, redisV1, redisV3, elasticEU, elasticUS},
			}

			externalServices := &mesh_core.ExternalServiceResourceList{}

			// when
			targets := BuildEndpointMap(defaultMeshWithMTLS, "zone-1", dataplanes.Items, externalServices.Items, dataSourceLoader)

			Expect(targets).To(HaveLen(4))
			// and
			Expect(targets).To(HaveKeyWithValue("redis", []core_xds.Endpoint{
				{
					Target: "192.168.0.2",
					Port:   6379,
					Tags:   map[string]string{"kuma.io/service": "redis", "version": "v1"},
					Weight: 1,
				},
				{
					Target: "192.168.0.4",
					Port:   6379,
					Tags: map[string]string{
						"kuma.io/service": "redis",
						"version":         "v3",
					},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("elastic", []core_xds.Endpoint{
				{
					Target: "192.168.0.5",
					Port:   9200,
					Tags: map[string]string{
						"kuma.io/service": "elastic",
						"region":          "eu",
					},
					Weight: 1,
				},
				{
					Target: "192.168.0.6",
					Port:   9200,
					Tags:   map[string]string{"kuma.io/service": "elastic", "region": "us"},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("frontend", []core_xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   7070,
					Tags: map[string]string{
						"kuma.io/service": "frontend",
						"region":          "eu",
					},
					Weight: 1,
				},
			}))
			Expect(targets).To(HaveKeyWithValue("backend", []core_xds.Endpoint{
				{
					Target: "192.168.0.1",
					Port:   8080,
					Tags: map[string]string{
						"kuma.io/service": "backend",
						"region":          "eu",
					},
					Weight: 1,
				},
			}))
		})
	})

	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			dataplanes       []*mesh_core.DataplaneResource
			externalServices []*mesh_core.ExternalServiceResource
			mesh             *mesh_core.MeshResource
			expected         core_xds.EndpointMap
		}
		DescribeTable("should include only those dataplanes that match given selectors",
			func(given testCase) {
				// when
				endpoints := BuildEndpointMap(given.mesh, "zone-1", given.dataplanes, given.externalServices, dataSourceLoader)
				// then
				Expect(endpoints).To(Equal(given.expected))
			},
			Entry("no dataplanes", testCase{
				dataplanes: []*mesh_core.DataplaneResource{},
				mesh:       defaultMeshWithMTLS,
				expected:   core_xds.EndpointMap{},
			}),
			Entry("ingress in the list of dataplanes", testCase{
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"kuma.io/service": "redis", "version": "v1"},
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
								Address: "10.20.1.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags: map[string]string{mesh_proto.ServiceTag: "ingress", mesh_proto.ZoneTag: "zone-2"},
										Port: 10001,
									},
								},
								Ingress: &mesh_proto.Dataplane_Networking_Ingress{
									AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
										{
											Instances: 2,
											Mesh:      defaultMeshName,
											Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2", "region": "eu"},
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
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v1"},
							Weight: 1,
						},
						{
							Target: "10.20.1.2",
							Port:   10001,
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v2", "region": "eu"},
							Weight: 2,
						},
						{
							Target: "10.20.1.2",
							Port:   10001,
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v3"},
							Weight: 3,
						},
					},
				},
			}),
			Entry("ingresses in the list of dataplanes from different meshes", testCase{
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"kuma.io/service": "redis", "version": "v1"},
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
								Address: "10.20.1.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags: map[string]string{mesh_proto.ServiceTag: "ingress", mesh_proto.ZoneTag: "zone-2"},
										Port: 10001,
									},
								},
								Ingress: &mesh_proto.Dataplane_Networking_Ingress{
									AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
										{
											Instances: 2,
											Mesh:      defaultMeshName,
											Tags:      map[string]string{"kuma.io/service": "redis", "version": "v2", "region": "eu"},
										},
										{
											Instances: 3,
											Mesh:      nonDefaultMesh,
											Tags:      map[string]string{"kuma.io/service": "redis", "version": "v3"},
										},
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
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v1"},
							Weight: 1,
						},
						{
							Target: "10.20.1.2",
							Port:   10001,
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v2", "region": "eu"},
							Weight: 2,
						},
					},
				},
			}),
			Entry("ingress is not included if mtls is off", testCase{
				dataplanes: []*mesh_core.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags:        map[string]string{"kuma.io/service": "redis", "version": "v1"},
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
								Address: "10.20.1.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Tags: map[string]string{mesh_proto.ServiceTag: "ingress", mesh_proto.ZoneTag: "zone-2"},
										Port: 10001,
									},
								},
								Ingress: &mesh_proto.Dataplane_Networking_Ingress{
									AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
										{
											Instances: 2,
											Mesh:      defaultMeshName,
											Tags:      map[string]string{"kuma.io/service": "redis", "version": "v2", "region": "eu"},
										},
										{
											Instances: 3,
											Mesh:      nonDefaultMesh,
											Tags:      map[string]string{"kuma.io/service": "redis", "version": "v3"},
										},
									},
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
							Tags:   map[string]string{"kuma.io/service": "redis", "version": "v1"},
							Weight: 1,
						},
					},
				},
			}),
			Entry("external service no TLS", testCase{
				dataplanes: []*mesh_core.DataplaneResource{},
				externalServices: []*mesh_core.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls:     nil,
							},
							Tags: map[string]string{"kuma.io/service": "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{"kuma.io/service": "redis"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service with TLS disabled", testCase{
				dataplanes: []*mesh_core.DataplaneResource{},
				externalServices: []*mesh_core.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									Enabled: false,
								},
							},
							Tags: map[string]string{"kuma.io/service": "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{"kuma.io/service": "redis"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: false},
						},
					},
				},
			}),
			Entry("external service with TLS enabled", testCase{
				dataplanes: []*mesh_core.DataplaneResource{},
				externalServices: []*mesh_core.ExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName},
						Spec: mesh_proto.ExternalService{
							Networking: &mesh_proto.ExternalService_Networking{
								Address: "httpbin.org:80",
								Tls: &mesh_proto.ExternalService_Networking_TLS{
									Enabled: true,
								},
							},
							Tags: map[string]string{"kuma.io/service": "redis"},
						},
					},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"redis": []core_xds.Endpoint{
						{
							Target:          "httpbin.org",
							Port:            80,
							Tags:            map[string]string{"kuma.io/service": "redis", "kuma.io/service_tls": "enabled"},
							Weight:          1,
							ExternalService: &core_xds.ExternalService{TLSEnabled: true},
						},
					},
				},
			}),
		)
	})
})
