package generator_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("InboundProxyGenerator", func() {

	type testCase struct {
		dataplaneFile string
		expected      string
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.InboundProxyGenerator{}
			ctx := xds_context.Context{
				ConnectionInfo: xds_context.ConnectionInfo{
					Authority: "kuma-system:5677",
				},
				ControlPlane: &xds_context.ControlPlaneContext{
					SdsTlsCert: []byte("12345"),
				},
				Mesh: xds_context.MeshContext{
					Resource: &mesh_core.MeshResource{
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
									},
								},
							},
						},
					},
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := ioutil.ReadFile(filepath.Join("testdata", "inbound-proxy", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			proxy := &model.Proxy{
				Id: model.ProxyId{Name: "side-car"},
				Dataplane: &mesh_core.DataplaneResource{
					Meta: &test_model.ResourceMeta{
						Version: "1",
					},
					Spec: &dataplane,
				},
				APIVersion: envoy_common.APIV3,
				Policies: model.MatchedPolicies{

					TrafficPermissions: model.TrafficPermissionMap{
						mesh_proto.InboundInterface{
							DataplaneIP:   "192.168.0.1",
							DataplanePort: 80,
							WorkloadIP:    "127.0.0.1",
							WorkloadPort:  8080,
						}: &mesh_core.TrafficPermissionResource{
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
							DataplaneIP:   "192.168.0.1",
							DataplanePort: 80,
							WorkloadIP:    "127.0.0.1",
							WorkloadPort:  8080,
						}: &mesh_proto.FaultInjection{
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
									Percentage: &wrappers.DoubleValue{Value: 50},
									Value:      &duration.Duration{Seconds: 5},
								},
							},
						},
					},
				},
				Metadata: &model.DataplaneMetadata{},
			}

			// when
			rs, err := gen.Generate(ctx, proxy)

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
	)
})
