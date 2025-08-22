package generator_test

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	test_xds "github.com/kumahq/kuma/pkg/test/xds"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("ProxyTemplateProfileSource", func() {
	type testCase struct {
		mesh            string
		dataplane       string
		profile         string
		expected        string
		features        xds_types.Features
		meshServiceMode mesh_proto.Mesh_MeshServices_Mode
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.ProxyTemplateProfileSource{
				ProfileName: given.profile,
			}

			// given
			outboundTargets := core_xds.EndpointMap{
				"db": []core_xds.Endpoint{
					{
						Target: "192.168.0.3",
						Port:   5432,
						Tags:   map[string]string{"kuma.io/service": "db", "role": "master"},
						Weight: 1,
					},
				},
				"elastic": []core_xds.Endpoint{
					{
						Target: "192.168.0.4",
						Port:   9200,
						Tags:   map[string]string{"kuma.io/service": "elastic"},
						Weight: 1,
					},
				},
			}
			resources := xds_context.NewResources()
			resources.MeshLocalResources[core_mesh.TrafficRouteType] = &core_mesh.TrafficRouteResourceList{
				Items: []*core_mesh.TrafficRouteResource{{
					Meta: &test_model.ResourceMeta{Name: "default-allow-all"},
					Spec: &mesh_proto.TrafficRoute{
						Sources: []*mesh_proto.Selector{{
							Match: mesh_proto.MatchAnyService(),
						}},
						Destinations: []*mesh_proto.Selector{{
							Match: mesh_proto.MatchAnyService(),
						}},
						Conf: &mesh_proto.TrafficRoute_Conf{
							Destination: mesh_proto.MatchAnyService(),
							LoadBalancer: &mesh_proto.TrafficRoute_LoadBalancer{
								LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
							},
						},
					},
				}},
			}
			ctx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					CLACache: &test_xds.DummyCLACache{OutboundTargets: outboundTargets},
					Secrets:  &test_xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "demo",
						},
						Spec: &mesh_proto.Mesh{
							MeshServices: &mesh_proto.Mesh_MeshServices{
								Mode: given.meshServiceMode,
							},
						},
					},
					Resources: resources,
					ServicesInformation: map[string]*xds_context.ServiceInformation{
						"db": {
							TLSReadiness: true,
							Protocol:     core_meta.ProtocolUnknown,
						},
						"elastic": {
							TLSReadiness: true,
							Protocol:     core_meta.ProtocolUnknown,
						},
					},
				},
			}

			Expect(util_proto.FromYAML([]byte(given.mesh), ctx.Mesh.Resource.Spec)).To(Succeed())

			dataplane := &mesh_proto.Dataplane{}
			Expect(util_proto.FromYAML([]byte(given.dataplane), dataplane)).To(Succeed())

			proxy := &core_xds.Proxy{
				Id: *core_xds.BuildProxyId("demo", "backend-01"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Name:    "backend-01",
						Mesh:    "demo",
						Version: "1",
					},
					Spec: dataplane,
				},
				SecretsTracker: envoy_common.NewSecretsTracker("demo", []string{"demo"}),
				APIVersion:     envoy_common.APIV3,
				Routing: core_xds.Routing{
					TrafficRoutes: core_xds.RouteMap{
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 54321,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("db"),
								},
							},
						},
						mesh_proto.OutboundInterface{
							DataplaneIP:   "127.0.0.1",
							DataplanePort: 59200,
						}: &core_mesh.TrafficRouteResource{
							Spec: &mesh_proto.TrafficRoute{
								Conf: &mesh_proto.TrafficRoute_Conf{
									Destination: mesh_proto.MatchService("elastic"),
								},
							},
						},
					},
					OutboundTargets: outboundTargets,
				},
				Policies: core_xds.MatchedPolicies{
					HealthChecks: core_xds.HealthCheckMap{
						"elastic": &core_mesh.HealthCheckResource{
							Spec: &mesh_proto.HealthCheck{
								Sources: []*mesh_proto.Selector{
									{Match: mesh_proto.TagSelector{"kuma.io/service": "*"}},
								},
								Destinations: []*mesh_proto.Selector{
									{Match: mesh_proto.TagSelector{"kuma.io/service": "elastic"}},
								},
								Conf: &mesh_proto.HealthCheck_Conf{
									Interval:           util_proto.Duration(5 * time.Second),
									Timeout:            util_proto.Duration(4 * time.Second),
									UnhealthyThreshold: 3,
									HealthyThreshold:   2,
								},
							},
						},
					},
				},
				Metadata: &core_xds.DataplaneMetadata{
					AdminPort:     9902,
					ReadinessPort: 9903,
					Version: &mesh_proto.Version{
						KumaDp: &mesh_proto.KumaDpVersion{
							Version: "1.2.0",
						},
					},
					WorkDir:  "/tmp",
					Features: given.features,
				},
				EnvoyAdminMTLSCerts: core_xds.ServerSideMTLSCerts{
					CaPEM: []byte("caPEM"),
					ServerPair: tls.KeyPair{
						CertPEM: []byte("certPEM"),
						KeyPEM:  []byte("keyPEM"),
					},
				},
				InternalAddresses: []core_xds.InternalAddress{
					{AddressPrefix: "10.8.0.0", PrefixLen: 16},
				},
			}

			// when
			rs, err := gen.Generate(context.Background(), core_xds.NewResourceSet(), ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			resp, err := rs.List().ToDeltaDiscoveryResponse()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(resp)
			// then
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files

			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "profile-source", given.expected)))
		},
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; readiness with TCP port 9903", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 59200
                tags:
                  kuma.io/service: elastic
`,
			profile:  core_mesh.ProfileDefaultProxy,
			expected: "1-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; readiness with TCP port 9903", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 59200
                tags:
                  kuma.io/service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
                ipFamilyMode: IPv4
`,
			profile:  core_mesh.ProfileDefaultProxy,
			expected: "2-envoy-config.golden.yaml",
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; prometheus_metrics=true; readiness with Unix socket", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /non-standard-path
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 59200
                tags:
                  kuma.io/service: elastic
`,
			profile:  core_mesh.ProfileDefaultProxy,
			expected: "3-envoy-config.golden.yaml",
			features: map[string]bool{
				xds_types.FeatureReadinessUnixSocket: true,
			},
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=true; prometheus_metrics=true; readiness with Unix socket", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
            metrics:
              enabledBackend: prometheus-1
              backends:
              - name: prometheus-1
                type: prometheus
                conf:
                  port: 1234
                  path: /foo/bar
                  skipMTLS: false
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
                    kuma.io/protocol: http
              outbound:
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 59200
                tags:
                  kuma.io/service: elastic
              transparentProxying:
                redirectPortOutbound: 15001
                redirectPortInbound: 15006
                ipFamilyMode: IPv4
`,
			profile:  core_mesh.ProfileDefaultProxy,
			expected: "4-envoy-config.golden.yaml",
			features: map[string]bool{
				xds_types.FeatureReadinessUnixSocket: true,
			},
		}),
		Entry("should support pre-defined `default-proxy` profile; transparent_proxying=false; unified naming; readiness with Unix socket", testCase{
			mesh: `
            mtls:
              enabledBackend: builtin
              backends:
              - type: builtin
                name: builtin
`,
			dataplane: `
            networking:
              address: 192.168.0.1
              inbound:
                - port: 80
                  servicePort: 8080
                  tags:
                    kuma.io/service: backend
              outbound:
              - port: 54321
                tags:
                  kuma.io/service: db
              - port: 59200
                tags:
                  kuma.io/service: elastic
`,
			profile:  core_mesh.ProfileDefaultProxy,
			meshServiceMode: mesh_proto.Mesh_MeshServices_Exclusive,
			expected: "5-envoy-config.golden.yaml",
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
				xds_types.FeatureReadinessUnixSocket:   true,
			},
		}),
	)
})
