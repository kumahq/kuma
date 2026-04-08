package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	. "github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/tls"
	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	"github.com/kumahq/kuma/v2/pkg/xds/generator"
)

var _ = Describe("AdminProxyGenerator", func() {
	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile    string
		expected         string
		adminAddress     string
		adminSocketPath  string
		features         xds_types.Features
		meshServicesMode mesh_proto.Mesh_MeshServices_Mode
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
						Spec: &mesh_proto.Mesh{
							MeshServices: &mesh_proto.Mesh_MeshServices{
								Mode: given.meshServicesMode,
							},
						},
					},
				},
			}

			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:       9901,
					AdminAddress:    given.adminAddress,
					AdminSocketPath: given.adminSocketPath,
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
		Entry("should generate admin resources, empty admin address", testCase{
			dataplaneFile: "01.dataplane.input.yaml",
			expected:      "01.envoy-config.golden.yaml",
			adminAddress:  "",
		}),
		Entry("should generate admin resources, IPv4 loopback", testCase{
			dataplaneFile: "02.dataplane.input.yaml",
			expected:      "02.envoy-config.golden.yaml",
			adminAddress:  "127.0.0.1",
		}),
		Entry("should generate admin resources, IPv6 loopback", testCase{
			dataplaneFile: "03.dataplane.input.yaml",
			expected:      "03.envoy-config.golden.yaml",
			adminAddress:  "::1",
		}),
		Entry("should generate admin resources, unspecified IPv4", testCase{
			dataplaneFile: "04.dataplane.input.yaml",
			expected:      "04.envoy-config.golden.yaml",
			adminAddress:  "0.0.0.0",
		}),
		Entry("should generate admin resources, unspecified IPv6", testCase{
			dataplaneFile: "05.dataplane.input.yaml",
			expected:      "05.envoy-config.golden.yaml",
			adminAddress:  "::",
		}),
		Entry("should generate admin resources, IPv6", testCase{
			dataplaneFile: "06.dataplane.input.yaml",
			expected:      "06.envoy-config.golden.yaml",
			adminAddress:  "::1",
		}),
		Entry("should generate admin resources, unified naming", testCase{
			dataplaneFile:    "07.dataplane.input.yaml",
			expected:         "07.envoy-config.golden.yaml",
			adminAddress:     "",
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
		}),
		Entry("should generate admin resources, unified naming, readiness via Unix socket", testCase{
			dataplaneFile:    "08.dataplane.input.yaml",
			expected:         "08.envoy-config.golden.yaml",
			adminAddress:     "127.0.0.1",
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
			features: map[string]bool{
				xds_types.FeatureUnifiedResourceNaming: true,
			},
		}),
		Entry("should generate admin resources, admin with Unix socket", testCase{
			dataplaneFile:    "09.dataplane.input.yaml",
			expected:         "09.envoy-config.golden.yaml",
			adminAddress:     "127.0.0.1",
			adminSocketPath:  "/tmp/kuma-dp/kuma-envoy-admin.sock",
			meshServicesMode: mesh_proto.Mesh_MeshServices_Exclusive,
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
					AdminPort:    9901,
					AdminAddress: given.adminAddress,
					Features:     given.features,
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
			expected:     `envoy admin cluster is not allowed to have addresses other than "", "0.0.0.0", "127.0.0.1", "::", "::1"`,
			adminAddress: "192.168.0.1", // it's not allowed to use such address
		}),
	)
})
