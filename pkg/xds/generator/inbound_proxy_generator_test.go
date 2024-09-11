package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("InboundProxyGenerator", func() {
	type testCase struct {
		dataplaneFile string
		expected      string
		mode          mesh_proto.CertificateAuthorityBackend_Mode
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.InboundProxyGenerator{}
			xdsCtx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{
							Mtls: &mesh_proto.Mesh_Mtls{
								EnabledBackend: "builtin",
								Backends: []*mesh_proto.CertificateAuthorityBackend{
									{
										Name: "builtin",
										Type: "builtin",
										Mode: given.mode,
									},
								},
							},
						},
					},
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "inbound-proxy", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy := &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: &dataplane,
				},
				SecretsTracker: envoy_common.NewSecretsTracker(xdsCtx.Mesh.Resource.Meta.GetName(), []string{xdsCtx.Mesh.Resource.Meta.GetName()}),
				APIVersion:     envoy_common.APIV3,
				Policies: model.MatchedPolicies{
					TrafficPermissions: model.TrafficPermissionMap{
						mesh_proto.InboundInterface{
							DataplaneAdvertisedIP: "192.168.0.1",
							DataplaneIP:           "192.168.0.1",
							DataplanePort:         80,
							WorkloadIP:            "192.168.0.1",
							WorkloadPort:          8080,
						}: &core_mesh.TrafficPermissionResource{
							Meta: &test_model.ResourceMeta{
								Name: "tp-1",
								Mesh: "default",
							},
							Spec: &mesh_proto.TrafficPermission{
								Sources: []*mesh_proto.Selector{
									{
										Match: map[string]string{
											"kuma.io/service": "web1",
											"version":         "1.0",
										},
									},
								},
								Destinations: []*mesh_proto.Selector{
									{
										Match: map[string]string{
											"kuma.io/service": "backend1",
											"env":             "dev",
										},
									},
								},
							},
						},
					},
					FaultInjections: model.FaultInjectionMap{
						mesh_proto.InboundInterface{
							DataplaneAdvertisedIP: "192.168.0.1",
							DataplaneIP:           "192.168.0.1",
							DataplanePort:         80,
							WorkloadIP:            "192.168.0.1",
							WorkloadPort:          8080,
						}: []*core_mesh.FaultInjectionResource{{Spec: &mesh_proto.FaultInjection{
							Sources: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "frontend",
									},
								},
							},
							Destinations: []*mesh_proto.Selector{
								{
									Match: map[string]string{
										"kuma.io/service": "backend1",
									},
								},
							},
							Conf: &mesh_proto.FaultInjection_Conf{
								Delay: &mesh_proto.FaultInjection_Conf_Delay{
									Percentage: util_proto.Double(50),
									Value:      util_proto.Duration(time.Second * 5),
								},
							},
						}}},
					},
					RateLimitsInbound: model.InboundRateLimitsMap{
						mesh_proto.InboundInterface{
							DataplaneAdvertisedIP: "192.168.0.1",
							DataplaneIP:           "192.168.0.1",
							DataplanePort:         80,
							WorkloadIP:            "192.168.0.1",
							WorkloadPort:          8080,
						}: []*core_mesh.RateLimitResource{
							{
								Spec: &mesh_proto.RateLimit{
									Sources: []*mesh_proto.Selector{
										{
											Match: map[string]string{
												"kuma.io/service": "frontend",
											},
										},
									},
									Destinations: []*mesh_proto.Selector{
										{
											Match: map[string]string{
												"kuma.io/service": "backend1",
											},
										},
									},
									Conf: &mesh_proto.RateLimit_Conf{
										Http: &mesh_proto.RateLimit_Conf_Http{
											Requests: 200,
											Interval: util_proto.Duration(time.Second * 10),
										},
									},
								},
							},
							{
								Spec: &mesh_proto.RateLimit{
									Sources: []*mesh_proto.Selector{
										{
											Match: map[string]string{
												"kuma.io/service": "*",
											},
										},
									},
									Destinations: []*mesh_proto.Selector{
										{
											Match: map[string]string{
												"kuma.io/service": "backend1",
											},
										},
									},
									Conf: &mesh_proto.RateLimit_Conf{
										Http: &mesh_proto.RateLimit_Conf_Http{
											Requests: 100,
											Interval: util_proto.Duration(time.Second * 2),
											OnRateLimit: &mesh_proto.RateLimit_Conf_Http_OnRateLimit{
												Status: util_proto.UInt32(404),
												Headers: []*mesh_proto.RateLimit_Conf_Http_OnRateLimit_HeaderValue{
													{
														Key:    "x-rate-limited",
														Value:  "true",
														Append: util_proto.Bool(false),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{},
			}

			// when
			rs, err := gen.Generate(context.Background(), nil, xdsCtx, proxy)

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
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "inbound-proxy", given.expected)))
		},
		Entry("01. transparent_proxying=false, ip_addresses=0, ports=0", testCase{
			dataplaneFile: "1-dataplane.input.yaml",
			expected:      "1-envoy-config.golden.yaml",
		}),
		Entry("02. transparent_proxying=true, ip_addresses=0, ports=0", testCase{
			dataplaneFile: "2-dataplane.input.yaml",
			expected:      "2-envoy-config.golden.yaml",
		}),
		Entry("03. transparent_proxying=false, ip_addresses=2, ports=2", testCase{
			dataplaneFile: "3-dataplane.input.yaml",
			expected:      "3-envoy-config.golden.yaml",
		}),
		Entry("04. transparent_proxying=true, ip_addresses=2, ports=2", testCase{
			dataplaneFile: "4-dataplane.input.yaml",
			expected:      "4-envoy-config.golden.yaml",
		}),
		Entry("05. transparent_proxying=false, ip_addresses=2, ports=2, mode=permissive", testCase{
			dataplaneFile: "5-dataplane.input.yaml",
			expected:      "5-envoy-config.golden.yaml",
			mode:          mesh_proto.CertificateAuthorityBackend_PERMISSIVE,
		}),
		Entry("06. transparent_proxying=true, ip_addresses=2, ports=2, mode=permissive", testCase{
			dataplaneFile: "6-dataplane.input.yaml",
			expected:      "6-envoy-config.golden.yaml",
			mode:          mesh_proto.CertificateAuthorityBackend_PERMISSIVE,
		}),
		Entry("07. transparent_proxying=true, ip_addresses=2, ports=2, mode=strict", testCase{
			dataplaneFile: "7-dataplane.input.yaml",
			expected:      "7-envoy-config.golden.yaml",
			mode:          mesh_proto.CertificateAuthorityBackend_STRICT,
		}),
	)
})
