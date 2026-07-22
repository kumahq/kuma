package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/tls"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("AdminProxyGenerator", func() {
	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile   string
		expected        string
		adminAddress    string
		adminSocketPath string
		readinessPort   uint32
		features        xds_types.Features
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := os.ReadFile(filepath.Join("testdata", "admin", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
			}

			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:       9901,
					AdminAddress:    given.adminAddress,
					AdminSocketPath: given.adminSocketPath,
					ReadinessPort:   given.readinessPort,
					Features:        given.features,
					IPv6Enabled:     true,
				},
				EnvoyAdminMTLSCerts: xds.ServerSideMTLSCerts{
					CaPEM: []byte("caPEM"),
					ServerPair: tls.KeyPair{
						CertPEM: []byte("certPEM"),
						KeyPEM:  []byte("keyPEM"),
					},
				},
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV3,
			}

			// when
			resources, err := generator.Generate(context.Background(), nil, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "admin", given.expected)))
		},
		Entry("should generate admin resources, empty admin address, readiness with TCP port 9902", testCase{
			dataplaneFile: "01.dataplane.input.yaml",
			expected:      "01.envoy-config.golden.yaml",
			adminAddress:  "",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, IPv4 loopback, readiness with TCP port 9902", testCase{
			dataplaneFile: "02.dataplane.input.yaml",
			expected:      "02.envoy-config.golden.yaml",
			adminAddress:  "127.0.0.1",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, IPv6 loopback, readiness with TCP port 9902", testCase{
			dataplaneFile: "03.dataplane.input.yaml",
			expected:      "03.envoy-config.golden.yaml",
			adminAddress:  "::1",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, unspecified IPv4, readiness with TCP port 9902", testCase{
			dataplaneFile: "04.dataplane.input.yaml",
			expected:      "04.envoy-config.golden.yaml",
			adminAddress:  "0.0.0.0",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, unspecified IPv6, readiness with TCP port 9902", testCase{
			dataplaneFile: "05.dataplane.input.yaml",
			expected:      "05.envoy-config.golden.yaml",
			adminAddress:  "::",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, Unix socket disabled, IPv6 with readiness with TCP port 9400", testCase{
			dataplaneFile: "06.dataplane.input.yaml",
			expected:      "06.envoy-config.golden.yaml",
			adminAddress:  "::1",
			readinessPort: 9400,
		}),
		Entry("should generate admin resources, unified naming, readiness with TCP port 9902", testCase{
			dataplaneFile: "07.dataplane.input.yaml",
			expected:      "07.envoy-config.golden.yaml",
			adminAddress:  "",
			readinessPort: 9902,
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
		}),
		Entry("should generate admin resources, legacy DP advertising readiness Unix socket", testCase{
			dataplaneFile: "08.dataplane.input.yaml",
			expected:      "08.envoy-config.golden.yaml",
			adminAddress:  "127.0.0.1",
			readinessPort: 9902,
			features: map[string]bool{
				xds_types.FeatureReadinessUnixSocket: true,
			},
		}),
		Entry("should generate admin resources, admin with Unix socket", testCase{
			dataplaneFile:   "09.dataplane.input.yaml",
			expected:        "09.envoy-config.golden.yaml",
			adminAddress:    "127.0.0.1",
			adminSocketPath: "/tmp/kuma-dp/kuma-envoy-admin.sock",
			readinessPort:   9902,
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
		}),
	)

	DescribeTable("should return error",
		func(given testCase) {
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}
			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:     9901,
					AdminAddress:  given.adminAddress,
					ReadinessPort: given.readinessPort,
					Features:      given.features,
				},
				EnvoyAdminMTLSCerts: xds.ServerSideMTLSCerts{
					CaPEM: []byte("caPEM"),
					ServerPair: tls.KeyPair{
						CertPEM: []byte("certPEM"),
						KeyPEM:  []byte("keyPEM"),
					},
				},
				Dataplane:  core_mesh.NewDataplaneResource(),
				APIVersion: envoy_common.APIV3,
				// internal addresses are set to "localhost" addresses to the "admin" listener
				// because user set x-envoy headers do not apply to this listener
				// we are settings these values here to assert they should not be generated onto Envoy config of the listener
				InternalAddresses: []xds.InternalAddress{
					{AddressPrefix: "10.0.0.0", PrefixLen: 8},
					{AddressPrefix: "127.0.0.1", PrefixLen: 32},
				},
			}

			// when
			_, err := generator.Generate(context.Background(), nil, ctx, proxy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.expected))
		},
		Entry("should return error when admin address is not allowed", testCase{
			expected:      `envoy admin cluster is not allowed to have addresses other than "", "0.0.0.0", "127.0.0.1", "::", "::1"`,
			adminAddress:  "192.168.0.1", // it's not allowed to use such address
			readinessPort: 9902,
		}),
		Entry("should return error when readiness port is 0", testCase{
			expected:      "ReadinessPort has to be in (0, 65535] range",
			adminAddress:  "127.0.0.1",
			readinessPort: 0,
		}),
	)

	// Zone ingress/egress aren't mesh-scoped, so xdsCtx.Mesh.Resource is always nil
	// for them. AdminProxyGenerator must still honor the unified naming feature flag
	// in that case instead of silently falling back to legacy naming.
	DescribeTable("should name the envoy admin cluster for zone proxies based on the unified naming feature",
		func(unifiedNaming bool, expectedClusterName string) {
			metadata := &xds.DataplaneMetadata{
				AdminPort:     9901,
				ReadinessPort: 9902,
				Features: map[string]bool{
					xds_types.FeatureUnifiedResourceNaming: unifiedNaming,
				},
			}
			ctx := xds_context.Context{}

			zoneIngressProxy := &xds.Proxy{
				Id:         *xds.BuildProxyId("default", "zone-ingress"),
				APIVersion: envoy_common.APIV3,
				ZoneIngressProxy: &xds.ZoneIngressProxy{
					ZoneIngressResource: &core_mesh.ZoneIngressResource{
						Meta: &test_model.ResourceMeta{Name: "zone-ingress"},
						Spec: &mesh_proto.ZoneIngress{
							Networking: &mesh_proto.ZoneIngress_Networking{Address: "10.0.0.1"},
						},
					},
				},
				Metadata: metadata,
			}
			zoneEgressProxy := &xds.Proxy{
				Id:         *xds.BuildProxyId("default", "zone-egress"),
				APIVersion: envoy_common.APIV3,
				ZoneEgressProxy: &xds.ZoneEgressProxy{
					ZoneEgressResource: &core_mesh.ZoneEgressResource{
						Meta: &test_model.ResourceMeta{Name: "zone-egress"},
						Spec: &mesh_proto.ZoneEgress{
							Networking: &mesh_proto.ZoneEgress_Networking{Address: "10.0.0.2"},
						},
					},
				},
				Metadata: metadata,
			}

			for _, proxy := range []*xds.Proxy{zoneIngressProxy, zoneEgressProxy} {
				resources, err := generator.Generate(context.Background(), nil, ctx, proxy)
				Expect(err).ToNot(HaveOccurred())

				var clusterNames []string
				for _, resource := range resources.List() {
					clusterNames = append(clusterNames, resource.Name)
				}
				Expect(clusterNames).To(ContainElement(expectedClusterName))
			}
		},
		Entry("legacy naming", false, "kuma:envoy:admin"),
		Entry("unified naming", true, "system_envoy_admin"),
	)
})
