package generator_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ = Describe("AdminProxyGenerator", func() {
	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile string
		expected      string
		adminAddress  string
		readinessPort uint32
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
					},
				},
			}

			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:     9901,
					AdminAddress:  given.adminAddress,
					ReadinessPort: given.readinessPort,
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
		Entry("should generate admin resources, unspecified IPv4, readiness port 0", testCase{
			dataplaneFile: "04.dataplane.input.yaml",
			expected:      "04.envoy-config.golden.yaml",
			adminAddress:  "0.0.0.0",
			readinessPort: 0,
		}),
		Entry("should generate admin resources, unspecified IPv6", testCase{
			dataplaneFile: "05.dataplane.input.yaml",
			expected:      "05.envoy-config.golden.yaml",
			adminAddress:  "::",
		}),
		Entry("should generate admin resources, IPv4 with readiness port 9902", testCase{
			dataplaneFile: "04.dataplane.input.yaml",
			expected:      "06.envoy-config.golden.yaml",
			adminAddress:  "127.0.0.1",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, IPv6 with readiness port 9400", testCase{
			dataplaneFile: "05.dataplane.input.yaml",
			expected:      "07.envoy-config.golden.yaml",
			adminAddress:  "::1",
			readinessPort: 9400,
		}),
	)

	It("should return error when admin address is not allowed", func() {
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
				AdminAddress: "192.168.0.1", // it's not allowed to use such address
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
		}

		// when
		_, err := generator.Generate(context.Background(), nil, ctx, proxy)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`envoy admin cluster is not allowed to have addresses other than "", "0.0.0.0", "127.0.0.1", "::", "::1"`))
	})
})
