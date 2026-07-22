package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/v3/pkg/core/xds"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("InboundProxyGenerator", func() {
	type testCase struct {
		dataplaneFile         string
		dataplaneMeta         *model.DataplaneMetadata
		dataplaneResourceMeta *test_model.ResourceMeta
		expected              string
		mode                  mesh_proto.CertificateAuthorityBackend_Mode
		casByTrustDomain      map[string][]xds_context.PEMBytes
	}

	DescribeTable("Generate Envoy xDS resources",
		func(given testCase) {
			// setup
			gen := &generator.InboundProxyGenerator{}
			mesh := &core_mesh.MeshResource{
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
			}

			xdsCtx := xds_context.Context{
				ControlPlane: &xds_context.ControlPlaneContext{
					Secrets: &xds.TestSecrets{},
				},
				Mesh: xds_context.MeshContext{
					Resource:         mesh,
					CAsByTrustDomain: given.casByTrustDomain,
				},
			}

			dataplane := mesh_proto.Dataplane{}
			dpBytes, err := os.ReadFile(filepath.Join("testdata", "inbound-proxy", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			Expect(util_proto.FromYAML(dpBytes, &dataplane)).To(Succeed())
			dpMeta := given.dataplaneResourceMeta
			if dpMeta == nil {
				dpMeta = &test_model.ResourceMeta{Version: "1"}
			}
			proxy := &model.Proxy{
				Id: *model.BuildProxyId("", "side-car"),
				Dataplane: &core_mesh.DataplaneResource{
					Meta: dpMeta,
					Spec: &dataplane,
				},
				SecretsTracker: envoy_common.NewSecretsTracker(xdsCtx.Mesh.Resource.Meta.GetName(), []string{xdsCtx.Mesh.Resource.Meta.GetName()}),
				APIVersion:     envoy_common.APIV3,
				Metadata:       given.dataplaneMeta,
				InternalAddresses: []model.InternalAddress{
					{AddressPrefix: "100.64.0.0", PrefixLen: 16},
					{AddressPrefix: "fc00::/7", PrefixLen: 128},
					{AddressPrefix: "::1/128", PrefixLen: 128},
				},
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
		Entry("09. transparent_proxying=false, ip_addresses=2, ports=2, trust with old mesh mtls", testCase{
			dataplaneFile: "9-dataplane.input.yaml",
			expected:      "9-envoy-config.golden.yaml",
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"my-test.domain.com": {xds_context.PEMBytes("123")},
			},
		}),
		Entry("10. transparent_proxying=false, no kuma.io/service tag, http protocol", testCase{
			dataplaneFile: "10-dataplane.input.yaml",
			expected:      "10-envoy-config.golden.yaml",
		}),
		Entry("11. inbound with no tags, listener tags filled from Dataplane KRI", testCase{
			dataplaneFile: "11-dataplane.input.yaml",
			dataplaneResourceMeta: &test_model.ResourceMeta{
				Version: "1",
				Name:    "backend-7f9c",
				Mesh:    "default",
				Labels: map[string]string{
					mesh_proto.ZoneTag:          "east",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
					mesh_proto.DisplayName:      "backend-7f9c",
				},
			},
			expected: "11-envoy-config.golden.yaml",
		}),
	)
})
