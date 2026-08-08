package topology_test

import (
	"context"

	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datasource_api "github.com/kumahq/kuma/v3/api/common/v1alpha1/datasource"
	common_tls "github.com/kumahq/kuma/v3/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/datasource"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	. "github.com/kumahq/kuma/v3/pkg/xds/topology"
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
	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager := secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil, false)
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})
	Describe("BuildEndpointMap()", func() {
		type testCase struct {
			dataplanes           []*core_mesh.DataplaneResource
			meshServices         []*meshservice_api.MeshServiceResource
			meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource
			meshMultiZoneService []*meshmzservice_api.MeshMultiZoneServiceResource
			meshZoneAddresses    []*meshzoneaddress_api.MeshZoneAddressResource
			zoneEgressAddresses  []core_xds.ZoneEgressInstance
			mesh                 *core_mesh.MeshResource
			expected             core_xds.EndpointMap
		}
		DescribeTable("should include only those dataplanes that match given selectors",
			func(given testCase) {
				// when
				endpoints := BuildDataplaneEndpointMap(context.Background(), "zone-1", given.meshServices, given.meshMultiZoneService, given.meshExternalServices, given.dataplanes, given.meshZoneAddresses, dataSourceLoader, given.mesh.MTLSEnabled(), given.zoneEgressAddresses)
				// then
				Expect(endpoints).To(Equal(given.expected))
			},
			Entry("no dataplanes", testCase{
				dataplanes: []*core_mesh.DataplaneResource{},
				mesh:       defaultMeshWithMTLS,
				expected:   core_xds.EndpointMap{},
			}),
			Entry("unhealthy dataplane", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Name: "dp-1", Mesh: defaultMeshName, Labels: map[string]string{"app": "redis"}},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Name: "dp-2", Mesh: defaultMeshName, Labels: map[string]string{"app": "redis"}},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        6379,
										ServicePort: 16379,
										Health:      &mesh_proto.Dataplane_Networking_Inbound_Health{Ready: false},
									},
								},
							},
						},
					},
				},
				meshServices: []*meshservice_api.MeshServiceResource{
					builders.MeshService().
						WithName("redis").
						WithDataplaneLabelsSelectorKV("app", "redis").
						AddIntPort(6379, 6379, "tcp").
						Build(),
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"kri_msvc_default___redis_6379": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   6379,
							Tags:   map[string]string{"app": "redis"},
							Weight: 1,
						},
					},
				},
			}),
			Entry("uses MeshService", testCase{
				dataplanes: []*core_mesh.DataplaneResource{
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName, Name: "redis-0", Labels: map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"}},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.1",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        6379,
										ServicePort: 16379,
									},
								},
							},
						},
					},
					{
						Meta: &test_model.ResourceMeta{Mesh: defaultMeshName, Labels: map[string]string{"app": "kong"}},
						Spec: &mesh_proto.Dataplane{
							Networking: &mesh_proto.Dataplane_Networking{
								Address: "192.168.0.2",
								Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
									{
										Port:        80,
										ServicePort: 18080,
									},
									{
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
						WithDataplaneLabelsSelectorKV("app", "kong").
						AddIntPort(8080, 80, "http").
						AddIntPort(8081, 8001, "http").
						Build(),
					builders.MeshService().
						WithName("redis").
						WithDataplaneLabelsSelectorKV(mesh_proto.ServiceTag, "redis_svc_6379").
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
					"kri_msvc_default___redis_6379": []core_xds.Endpoint{
						{
							Target:   "192.168.0.1",
							Port:     6379,
							Tags:     map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"},
							Locality: nil,
							Weight:   1,
						},
					},
					"kri_msvc_default___redis-0_6379": []core_xds.Endpoint{
						{
							Target:   "192.168.0.1",
							Port:     6379,
							Tags:     map[string]string{mesh_proto.ServiceTag: "redis_svc_6379"},
							Locality: nil,
							Weight:   1,
						},
					},
					"kri_msvc_default___kong.kong-system_8080": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     80,
							Tags:     map[string]string{"app": "kong"},
							Locality: nil,
							Weight:   1,
						},
					},
					"kri_msvc_default___kong.kong-system_8081": []core_xds.Endpoint{
						{
							Target:   "192.168.0.2",
							Port:     8001,
							Tags:     map[string]string{"app": "kong"},
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
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolHTTP,
							},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    443,
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
									Mode:       meshexternalservice_api.TLSVerificationSecured,
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
									CaCert: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "ca"},
									},
									ClientCert: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "cert"},
									},
									ClientKey: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "key"},
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
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolTCP,
							},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    443,
								},
							},
							Tls: &meshexternalservice_api.Tls{
								Enabled: true,
								Verification: &meshexternalservice_api.Verification{
									Mode:       meshexternalservice_api.TLSVerificationSkipSAN,
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
				zoneEgressAddresses: []core_xds.ZoneEgressInstance{
					{Address: "1.1.1.1", Port: 10002},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"kri_extsvc_default___another-mes_10000": []core_xds.Endpoint{
						{
							Target: "1.1.1.1",
							Port:   10002,
							Tags: map[string]string{
								"custom-label": "label",
							},
							Locality: nil,
							Weight:   1,
							ExternalService: &core_xds.ExternalService{
								Protocol:                 core_meta.ProtocolTCP,
								TLSEnabled:               true,
								FallbackToSystemCa:       true,
								SkipHostnameVerification: true,
								SANs:                     []core_xds.SAN{},
								OwnerResource: kri.Identifier{
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
									Mesh:         "default",
									Name:         "another-mes",
								},
							},
						},
					},
					"kri_extsvc_default___example-mes_10000": []core_xds.Endpoint{
						{
							Target:   "1.1.1.1",
							Port:     10002,
							Locality: nil,
							Weight:   1,
							ExternalService: &core_xds.ExternalService{
								Protocol:                 core_meta.ProtocolHTTP,
								TLSEnabled:               true,
								FallbackToSystemCa:       true,
								CaCert:                   []byte("ca"),
								ClientCert:               []byte("cert"),
								ClientKey:                []byte("key"),
								AllowRenegotiation:       true,
								SkipHostnameVerification: false,
								ServerName:               "example.com",
								SANs: []core_xds.SAN{
									{
										MatchType: core_xds.SANMatchPrefix,
										Value:     "test.com",
									},
									{
										MatchType: core_xds.SANMatchExact,
										Value:     "test.com",
									},
								},
								MinTlsVersion: pointer.To(tlsv3.TlsParameters_TLSv1_2),
								MaxTlsVersion: pointer.To(tlsv3.TlsParameters_TLSv1_3),
								OwnerResource: kri.Identifier{
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
									Mesh:         "default",
									Name:         "example-mes",
								},
							},
						},
					},
				},
			}),
			Entry("skips MeshExternalService reading a control plane local file", testCase{
				meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "file-mes",
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolTCP,
							},
							// an extension owns tls validation, so File and EnvVar can only be
							// stopped here
							Extension: &meshexternalservice_api.Extension{Type: "example"},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    443,
								},
							},
							Tls: &meshexternalservice_api.Tls{
								Enabled: true,
								Verification: &meshexternalservice_api.Verification{
									Mode: meshexternalservice_api.TLSVerificationSecured,
									CaCert: &datasource_api.SecureDataSource{
										Type: datasource_api.SecureDataSourceFile,
										File: &datasource_api.File{Path: "/etc/hosts"},
									},
								},
							},
						},
					},
				},
				zoneEgressAddresses: []core_xds.ZoneEgressInstance{
					{Address: "1.1.1.1", Port: 10002},
				},
				mesh:     defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{},
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
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolHTTP,
							},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{
									Address: "example.com",
									Port:    443,
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
									Mode:       meshexternalservice_api.TLSVerificationSecured,
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
									CaCert: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "ca"},
									},
									ClientCert: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "cert"},
									},
									ClientKey: &datasource_api.SecureDataSource{
										Type:           datasource_api.SecureDataSourceInline,
										InsecureInline: &datasource_api.Inline{Value: "key"},
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
				meshZoneAddresses: []*meshzoneaddress_api.MeshZoneAddressResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:   defaultMeshName,
							Name:   "mza-east",
							Labels: map[string]string{mesh_proto.ZoneTag: "east"},
						},
						Spec: &meshzoneaddress_api.MeshZoneAddress{
							Address: "192.168.0.100",
							Port:    12345,
						},
					},
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
						AddMatchedMeshServiceName(kri.From(samples.MeshServiceBackend())).
						AddMatchedMeshServiceName(kri.From(samples.MeshServiceSyncedBackend())).
						Build(),
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"kri_msvc_default___backend_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   80,
							Tags: map[string]string{
								"kuma.io/workload": "backend",
							},
							Weight: 1,
						},
					},
					"kri_msvc_default_east__backend_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags: map[string]string{
								"kuma.io/service": "kri_msvc_default_east__backend_80",
								"kuma.io/zone":    "east",
							},
							Weight:   1,
							Locality: &core_xds.Locality{Zone: "east", SubZone: "", Priority: 1, Weight: 0},
						},
					},
					"kri_mzsvc_default___backend_80": []core_xds.Endpoint{
						{
							Target: "192.168.0.1",
							Port:   80,
							Tags: map[string]string{
								"kuma.io/workload": "backend",
							},
							Weight: 1,
						},
						{
							Target: "192.168.0.100",
							Port:   12345,
							Tags: map[string]string{
								"kuma.io/service": "kri_msvc_default_east__backend_80",
								"kuma.io/zone":    "east",
							},
							Weight:   1,
							Locality: &core_xds.Locality{Zone: "east", SubZone: "", Priority: 1, Weight: 0},
						},
					},
				},
			}),
			Entry("uses MeshExternalService with dataplane zone egress listener", testCase{
				meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "ext-svc",
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolTCP,
							},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{Address: "external.com", Port: 443},
							},
						},
					},
				},
				zoneEgressAddresses: []core_xds.ZoneEgressInstance{
					{Address: "10.42.0.11", Port: 10002},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"kri_extsvc_default___ext-svc_10000": []core_xds.Endpoint{
						{
							Target: "10.42.0.11",
							Port:   10002,
							Tags:   nil,
							Weight: 1,
							ExternalService: &core_xds.ExternalService{
								Protocol: core_meta.ProtocolTCP,
								OwnerResource: kri.Identifier{
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
									Mesh:         "default",
									Name:         "ext-svc",
								},
							},
						},
					},
				},
			}),
			Entry("prefers dataplane zone egress listener address", testCase{
				meshExternalServices: []*meshexternalservice_api.MeshExternalServiceResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh: "default",
							Name: "ext-svc",
						},
						Spec: &meshexternalservice_api.MeshExternalService{
							Match: meshexternalservice_api.Match{
								Type:     meshexternalservice_api.HostnameGeneratorType,
								Port:     10000,
								Protocol: core_meta.ProtocolTCP,
							},
							Endpoints: &[]meshexternalservice_api.Endpoint{
								{Address: "external.com", Port: 443},
							},
						},
					},
				},
				zoneEgressAddresses: []core_xds.ZoneEgressInstance{
					{Address: "10.42.0.11", Port: 10002},
				},
				mesh: defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{
					"kri_extsvc_default___ext-svc_10000": []core_xds.Endpoint{
						{
							Target: "10.42.0.11",
							Port:   10002,
							Tags:   nil,
							Weight: 1,
							ExternalService: &core_xds.ExternalService{
								Protocol: core_meta.ProtocolTCP,
								OwnerResource: kri.Identifier{
									ResourceType: meshexternalservice_api.MeshExternalServiceType,
									Mesh:         "default",
									Name:         "ext-svc",
								},
							},
						},
					},
				},
			}),
			Entry("remote MeshService without a MeshZoneAddress is not included", testCase{
				meshServices: []*meshservice_api.MeshServiceResource{
					samples.MeshServiceSyncedBackend(), // remote MeshService from "east" zone
				},
				mesh:     defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{},
			}),
			Entry("remote MeshService with a MeshZoneAddress of another zone is not included", testCase{
				meshZoneAddresses: []*meshzoneaddress_api.MeshZoneAddressResource{
					{
						Meta: &test_model.ResourceMeta{
							Mesh:   defaultMeshName,
							Name:   "mza-north",
							Labels: map[string]string{mesh_proto.ZoneTag: "north"},
						},
						Spec: &meshzoneaddress_api.MeshZoneAddress{
							Address: "192.168.0.100",
							Port:    12345,
						},
					},
				},
				meshServices: []*meshservice_api.MeshServiceResource{
					samples.MeshServiceSyncedBackend(), // remote MeshService from "east" zone
				},
				mesh:     defaultMeshWithMTLS,
				expected: core_xds.EndpointMap{},
			}),
		)
	})
})
